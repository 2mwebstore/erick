// Package services holds business logic, independent of transport and storage.
package services

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/kongchansila/portfolio/backend/internal/models"
)

// Field limits mirrored by frontend/utils/validation.ts. The server is
// authoritative; the client copy only saves a round trip.
const (
	NameMin    = 2
	NameMax    = 100
	EmailMax   = 254
	MessageMin = 10
	MessageMax = 4000
	PhoneMax   = 40
	SubjectMax = 160
)

// ProjectTypes is the closed set accepted by the form (§20).
var ProjectTypes = []string{
	"Web Application",
	"Mobile Application",
	"E-commerce",
	"Backend / API",
	"WordPress / SEO",
	"DevOps",
	"Other",
}

// ValidationError carries per-field messages safe to show a visitor.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %d field(s)", len(e.Fields))
}

// ContactRepository is the storage dependency, kept as an interface so the
// service is testable without a database.
type ContactRepository interface {
	Create(ctx context.Context, msg *models.ContactMessage) (int64, error)
}

type ContactService struct {
	repo ContactRepository
}

func NewContactService(repo ContactRepository) *ContactService {
	return &ContactService{repo: repo}
}

// ErrNoStorage is returned when the service has no repository configured.
var ErrNoStorage = errors.New("contact: no storage configured")

// Submit validates the request and persists it.
//
// Validation runs before storage so malformed or oversized input never reaches
// the database, and the returned error type tells the handler whether the fault
// was the client's (400) or ours (500).
func (s *ContactService) Submit(ctx context.Context, req models.ContactRequest, ip, userAgent string) (*models.ContactMessage, error) {
	msg, err := Normalise(req)
	if err != nil {
		return nil, err
	}

	msg.IPAddress = truncate(ip, 45) // fits an IPv6 literal
	msg.UserAgent = truncate(userAgent, 512)

	if s.repo == nil {
		return nil, ErrNoStorage
	}

	id, err := s.repo.Create(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("contact: persist: %w", err)
	}
	msg.ID = id

	return msg, nil
}

// Normalise trims, validates, and converts a request into a message.
// Exported so it can be tested directly.
func Normalise(req models.ContactRequest) (*models.ContactMessage, error) {
	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.Email)
	phone := strings.TrimSpace(req.Phone)
	subject := strings.TrimSpace(req.Subject)
	projectType := strings.TrimSpace(req.ProjectType)
	message := strings.TrimSpace(req.Message)

	fields := map[string]string{}

	switch n := utf8.RuneCountInString(name); {
	case n < NameMin:
		fields["name"] = "Please enter your name."
	case n > NameMax:
		fields["name"] = fmt.Sprintf("Keep this under %d characters.", NameMax)
	}

	switch {
	case email == "":
		fields["email"] = "Please enter your email address."
	case len(email) > EmailMax:
		fields["email"] = "That email address is too long."
	default:
		addr, err := mail.ParseAddress(email)
		// Reject display-name forms ("Name <a@b.c>") so what is stored is the
		// bare address and nothing else.
		if err != nil || addr.Address != email || !replyableDomain(email) {
			fields["email"] = "Please enter a valid email address."
		}
	}

	if projectType == "" {
		fields["project_type"] = "Please choose a project type."
	} else if !validProjectType(projectType) {
		fields["project_type"] = "Please choose a project type from the list."
	}

	switch n := utf8.RuneCountInString(message); {
	case n < MessageMin:
		fields["message"] = "Please add a few details about the project."
	case n > MessageMax:
		fields["message"] = fmt.Sprintf("Keep this under %d characters.", MessageMax)
	}

	// Both are optional — a message without them is still worth reading — so
	// only the upper bound is enforced.
	if len(phone) > PhoneMax {
		fields["phone"] = fmt.Sprintf("Keep this under %d characters.", PhoneMax)
	}
	if len(subject) > SubjectMax {
		fields["subject"] = fmt.Sprintf("Keep this under %d characters.", SubjectMax)
	}

	if !utf8.ValidString(name) || !utf8.ValidString(message) {
		fields["message"] = "Please use valid text characters."
	}

	if len(fields) > 0 {
		return nil, &ValidationError{Fields: fields}
	}

	return &models.ContactMessage{
		Name:        name,
		Email:       email,
		Phone:       phone,
		Subject:     subject,
		ProjectType: projectType,
		Message:     message,
	}, nil
}

// replyableDomain requires a dotted domain with an alphabetic TLD.
//
// net/mail accepts RFC-valid but unreplyable addresses such as "a@localhost".
// This form cannot be answered, and accepting it here would also disagree with
// the client-side check in frontend/utils/validation.ts — the two layers must
// reject the same set of inputs, or a visitor can be blocked by one and waved
// through by the other.
func replyableDomain(email string) bool {
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return false
	}

	domain := email[at+1:]
	dot := strings.LastIndex(domain, ".")
	if dot <= 0 || dot == len(domain)-1 {
		return false
	}

	tld := domain[dot+1:]
	if len(tld) < 2 {
		return false
	}
	for _, r := range tld {
		if r < 'a' || r > 'z' {
			if r < 'A' || r > 'Z' {
				return false
			}
		}
	}
	return true
}

func validProjectType(value string) bool {
	for _, t := range ProjectTypes {
		if t == value {
			return true
		}
	}
	return false
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
