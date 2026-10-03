package models

import "time"

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
)

// CanManageUsers reports whether the role may create or delete accounts.
// Editors can change content; only an admin can change who has access.
func (r Role) CanManageUsers() bool { return r == RoleAdmin }

func (r Role) Valid() bool { return r == RoleAdmin || r == RoleEditor }

type User struct {
	ID          int64      `json:"id"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	Role        Role       `json:"role"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`

	// PasswordHash never leaves the server; the json tag keeps it out of every
	// response by construction rather than by remembering to strip it.
	PasswordHash string `json:"-"`
}

type Session struct {
	TokenHash  string    `json:"-"`
	UserID     int64     `json:"user_id"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	LastSeenAt time.Time `json:"last_seen_at"`
	IPAddress  string    `json:"-"`
	UserAgent  string    `json:"-"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuditEntry struct {
	ID         int64     `json:"id"`
	UserID     *int64    `json:"user_id,omitempty"`
	ActorEmail string    `json:"actor_email"`
	Action     string    `json:"action"`
	Entity     string    `json:"entity"`
	EntityID   string    `json:"entity_id,omitempty"`
	Detail     any       `json:"detail,omitempty"`
	IPAddress  string    `json:"ip_address,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
