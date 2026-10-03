package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"github.com/kongchansila/portfolio/backend/internal/models"
	"github.com/kongchansila/portfolio/backend/internal/repositories"
)

var (
	// ErrInvalidCredentials is deliberately returned for both a missing account
	// and a wrong password: distinguishing them tells an attacker which emails
	// are registered.
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	ErrAccountDisabled    = errors.New("auth: account disabled")
	ErrSessionInvalid     = errors.New("auth: session invalid or expired")
)

const (
	// bcrypt cost 12: roughly 250ms per verification on modern hardware, which
	// is cheap for a handful of logins and expensive for an offline attacker.
	bcryptCost = 12

	SessionDuration = 12 * time.Hour
	sessionBytes    = 32
	PasswordMin     = 12
	PasswordMax     = 200
)

type AuthService struct {
	users    *repositories.UserRepository
	sessions *repositories.SessionRepository
}

func NewAuthService(users *repositories.UserRepository, sessions *repositories.SessionRepository) *AuthService {
	return &AuthService{users: users, sessions: sessions}
}

// Authenticated is what a successful login or session lookup yields.
type Authenticated struct {
	User    *models.User
	Token   string // only set on login; never stored
	Expires time.Time
}

// Login verifies credentials and issues a session.
func (s *AuthService) Login(ctx context.Context, email, password, ip, userAgent string) (*Authenticated, error) {
	user, err := s.users.ByEmail(ctx, email)

	if err != nil && !errors.Is(err, repositories.ErrNotFound) {
		return nil, err
	}

	if user == nil {
		// Spend comparable time on a missing account so response timing does not
		// reveal whether the email exists.
		_ = bcrypt.CompareHashAndPassword(
			[]byte("$2a$12$C6UzMDM.H6dfI/f/IKcEe.6qkVqDMNQBpuGi5ZBdVsUjMdCLgDFFm"),
			[]byte(password),
		)
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrAccountDisabled
	}

	token, hash, err := newSessionToken()
	if err != nil {
		return nil, err
	}

	expires := time.Now().UTC().Add(SessionDuration)
	if err := s.sessions.Create(ctx, &models.Session{
		TokenHash: hash,
		UserID:    user.ID,
		ExpiresAt: expires,
		IPAddress: ip,
		UserAgent: truncate(userAgent, 512),
	}); err != nil {
		return nil, err
	}

	if err := s.users.TouchLogin(ctx, user.ID); err != nil {
		// Not fatal: the login succeeded, only the bookkeeping failed.
		_ = err
	}

	return &Authenticated{User: user, Token: token, Expires: expires}, nil
}

// Resolve turns a session token from a cookie back into a user.
func (s *AuthService) Resolve(ctx context.Context, token string) (*models.User, error) {
	if token == "" {
		return nil, ErrSessionInvalid
	}

	_, user, err := s.sessions.Resolve(ctx, hashToken(token))
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrSessionInvalid
		}
		return nil, err
	}

	// Best-effort activity tracking; a failure here must not deny access.
	_ = s.sessions.Touch(ctx, hashToken(token))

	return user, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.sessions.Delete(ctx, hashToken(token))
}

// ChangePassword updates the hash and revokes every existing session, so a
// password change actually locks out whoever had the old one.
func (s *AuthService) ChangePassword(ctx context.Context, userID int64, current, next string) error {
	user, err := s.users.ByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(current)); err != nil {
		return ErrInvalidCredentials
	}

	if err := ValidatePassword(next); err != nil {
		return err
	}

	hash, err := HashPassword(next)
	if err != nil {
		return err
	}

	if err := s.users.UpdatePassword(ctx, userID, hash); err != nil {
		return err
	}
	return s.sessions.DeleteForUser(ctx, userID)
}

// CreateUser registers an account. Only an admin may call this.
func (s *AuthService) CreateUser(ctx context.Context, email, name, password string, role models.Role) (*models.User, error) {
	fields := map[string]string{}

	email = strings.ToLower(strings.TrimSpace(email))
	if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email {
		fields["email"] = "Please enter a valid email address."
	}
	if strings.TrimSpace(name) == "" {
		fields["name"] = "Please enter a name."
	}
	if !role.Valid() {
		fields["role"] = "Role must be admin or editor."
	}
	if err := ValidatePassword(password); err != nil {
		var invalid *ValidationError
		if errors.As(err, &invalid) {
			fields["password"] = invalid.Fields["password"]
		}
	}
	if len(fields) > 0 {
		return nil, &ValidationError{Fields: fields}
	}

	existing, err := s.users.ByEmail(ctx, email)
	if err != nil && !errors.Is(err, repositories.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, &ValidationError{Fields: map[string]string{"email": "That email is already registered."}}
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        email,
		Name:         strings.TrimSpace(name),
		PasswordHash: hash,
		Role:         role,
		IsActive:     true,
	}

	id, err := s.users.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	user.ID = id
	return user, nil
}

func (s *AuthService) ListUsers(ctx context.Context) ([]models.User, error) {
	return s.users.List(ctx)
}

func (s *AuthService) UserCount(ctx context.Context) (int, error) {
	return s.users.Count(ctx)
}

// HashPassword produces a bcrypt hash at the configured cost.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("auth: hash password: %w", err)
	}
	return string(hash), nil
}

// ValidatePassword enforces length and variety.
//
// Length does most of the work, so the floor is high (12) and the character
// rules are light — a long passphrase should not be rejected for lacking a
// symbol. bcrypt silently truncates at 72 bytes, so anything longer is refused
// rather than quietly shortened.
func ValidatePassword(password string) error {
	if len(password) > 72 {
		return &ValidationError{Fields: map[string]string{
			"password": "Passwords are limited to 72 characters.",
		}}
	}
	if len([]rune(password)) < PasswordMin {
		return &ValidationError{Fields: map[string]string{
			"password": fmt.Sprintf("Use at least %d characters.", PasswordMin),
		}}
	}

	var hasLetter, hasOther bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r), unicode.IsPunct(r), unicode.IsSymbol(r), unicode.IsSpace(r):
			hasOther = true
		}
	}
	if !hasLetter || !hasOther {
		return &ValidationError{Fields: map[string]string{
			"password": "Include letters plus at least one number, symbol or space.",
		}}
	}
	return nil
}

// newSessionToken returns the token to put in the cookie and the hash to store.
func newSessionToken() (token, hash string, err error) {
	raw := make([]byte, sessionBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("auth: generate session token: %w", err)
	}
	token = base64.RawURLEncoding.EncodeToString(raw)
	return token, hashToken(token), nil
}

// hashToken stores sessions by SHA-256 so a database leak yields no usable
// tokens. A fast hash is correct here: the input is 256 bits of entropy, not a
// guessable password.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// NewCSRFToken returns a random token for the double-submit cookie pattern.
func NewCSRFToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("auth: generate csrf token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// CSRFMatches compares in constant time, so a mismatch reveals nothing about
// how much of the token was correct.
func CSRFMatches(cookie, header string) bool {
	if cookie == "" || header == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookie), []byte(header)) == 1
}
