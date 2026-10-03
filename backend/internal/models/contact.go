// Package models holds the data structures that cross package boundaries.
package models

import "time"

// ContactMessage is one submission of the contact form.
//
// The request metadata is stored alongside the message so abusive traffic can
// be investigated after the fact without needing separate request logs.
type ContactMessage struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	// Optional on purpose: the form asks for them, but a message that arrives
	// without one is still a message worth reading.
	Phone       string    `json:"phone,omitempty"`
	Subject     string    `json:"subject,omitempty"`
	ProjectType string    `json:"project_type"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`

	// Read and archive state, managed from /admin.
	ReadAt     *time.Time `json:"read_at,omitempty"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`

	// Request metadata is exposed to the admin inbox for abuse investigation,
	// and never in a public response.
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// ContactRequest is the accepted request body for POST /v1/contact.
type ContactRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Subject     string `json:"subject"`
	ProjectType string `json:"project_type"`
	Message     string `json:"message"`
}
