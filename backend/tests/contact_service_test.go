package tests

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kongchansila/portfolio/backend/internal/models"
	"github.com/kongchansila/portfolio/backend/internal/services"
)

// stubRepo stands in for MySQL so validation and service behaviour can be tested
// without a database.
type stubRepo struct {
	created []*models.ContactMessage
	err     error
}

func (s *stubRepo) Create(_ context.Context, msg *models.ContactMessage) (int64, error) {
	if s.err != nil {
		return 0, s.err
	}
	s.created = append(s.created, msg)
	return int64(len(s.created)), nil
}

func validRequest() models.ContactRequest {
	return models.ContactRequest{
		Name:        "Test Visitor",
		Email:       "visitor@example.com",
		ProjectType: "Web Application",
		Message:     "We need an internal ordering system for about forty staff.",
	}
}

func TestNormaliseAcceptsValidRequest(t *testing.T) {
	msg, err := services.Normalise(validRequest())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if msg.Name != "Test Visitor" || msg.Email != "visitor@example.com" {
		t.Fatalf("unexpected normalised message: %+v", msg)
	}
}

func TestNormaliseTrimsWhitespace(t *testing.T) {
	req := validRequest()
	req.Name = "  Test Visitor  "
	req.Email = "  visitor@example.com "
	req.Message = "\n  A message with padding around it.  \n"

	msg, err := services.Normalise(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if msg.Name != "Test Visitor" {
		t.Errorf("name not trimmed: %q", msg.Name)
	}
	if msg.Email != "visitor@example.com" {
		t.Errorf("email not trimmed: %q", msg.Email)
	}
	if strings.HasPrefix(msg.Message, " ") || strings.HasSuffix(msg.Message, "\n") {
		t.Errorf("message not trimmed: %q", msg.Message)
	}
}

func TestNormaliseRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name  string
		mutit func(*models.ContactRequest)
		field string
	}{
		{"empty name", func(r *models.ContactRequest) { r.Name = "" }, "name"},
		{"short name", func(r *models.ContactRequest) { r.Name = "A" }, "name"},
		{"long name", func(r *models.ContactRequest) { r.Name = strings.Repeat("a", services.NameMax+1) }, "name"},
		{"empty email", func(r *models.ContactRequest) { r.Email = "" }, "email"},
		{"malformed email", func(r *models.ContactRequest) { r.Email = "not-an-email" }, "email"},
		{"email without tld", func(r *models.ContactRequest) { r.Email = "visitor@localhost" }, "email"},
		{"display name email", func(r *models.ContactRequest) { r.Email = "Visitor <v@example.com>" }, "email"},
		{"empty project type", func(r *models.ContactRequest) { r.ProjectType = "" }, "project_type"},
		{"unknown project type", func(r *models.ContactRequest) { r.ProjectType = "Blockchain" }, "project_type"},
		{"short message", func(r *models.ContactRequest) { r.Message = "hi" }, "message"},
		{"long message", func(r *models.ContactRequest) { r.Message = strings.Repeat("a", services.MessageMax+1) }, "message"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := validRequest()
			tc.mutit(&req)

			_, err := services.Normalise(req)
			if err == nil {
				t.Fatalf("expected a validation error for %s", tc.name)
			}

			var invalid *services.ValidationError
			if !errors.As(err, &invalid) {
				t.Fatalf("expected *services.ValidationError, got %T", err)
			}
			if _, ok := invalid.Fields[tc.field]; !ok {
				t.Fatalf("expected an error on %q, got %v", tc.field, invalid.Fields)
			}
		})
	}
}

func TestNormaliseCountsRunesNotBytes(t *testing.T) {
	req := validRequest()
	// Ten Khmer characters: ten runes, thirty bytes. A byte-length check would
	// wrongly accept this as meeting the minimum, or reject a valid long message.
	req.Message = strings.Repeat("ក", services.MessageMin-1)

	if _, err := services.Normalise(req); err == nil {
		t.Fatal("expected a message-too-short error for a 9-rune message")
	}

	req.Message = strings.Repeat("ក", services.MessageMin)
	if _, err := services.Normalise(req); err != nil {
		t.Fatalf("expected a 10-rune message to be accepted, got %v", err)
	}
}

func TestSubmitPersistsMessageWithMetadata(t *testing.T) {
	repo := &stubRepo{}
	svc := services.NewContactService(repo)

	msg, err := svc.Submit(context.Background(), validRequest(), "203.0.113.7", "Mozilla/5.0")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if msg.ID != 1 {
		t.Errorf("expected id 1, got %d", msg.ID)
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected 1 stored message, got %d", len(repo.created))
	}
	if repo.created[0].IPAddress != "203.0.113.7" {
		t.Errorf("ip not stored: %q", repo.created[0].IPAddress)
	}
	if repo.created[0].UserAgent != "Mozilla/5.0" {
		t.Errorf("user agent not stored: %q", repo.created[0].UserAgent)
	}
}

func TestSubmitDoesNotStoreInvalidInput(t *testing.T) {
	repo := &stubRepo{}
	svc := services.NewContactService(repo)

	req := validRequest()
	req.Email = "broken"

	if _, err := svc.Submit(context.Background(), req, "", ""); err == nil {
		t.Fatal("expected a validation error")
	}
	if len(repo.created) != 0 {
		t.Fatalf("invalid input reached storage: %+v", repo.created)
	}
}

func TestSubmitWrapsStorageFailure(t *testing.T) {
	sentinel := errors.New("connection refused")
	svc := services.NewContactService(&stubRepo{err: sentinel})

	_, err := svc.Submit(context.Background(), validRequest(), "", "")
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected the storage error to be wrapped, got %v", err)
	}

	var invalid *services.ValidationError
	if errors.As(err, &invalid) {
		t.Fatal("a storage failure must not be reported as a validation error")
	}
}

func TestSubmitWithoutStorageReturnsErrNoStorage(t *testing.T) {
	svc := services.NewContactService(nil)

	_, err := svc.Submit(context.Background(), validRequest(), "", "")
	if !errors.Is(err, services.ErrNoStorage) {
		t.Fatalf("expected ErrNoStorage, got %v", err)
	}
}

func TestTruncatesOverlongMetadata(t *testing.T) {
	repo := &stubRepo{}
	svc := services.NewContactService(repo)

	longAgent := strings.Repeat("u", 900)

	if _, err := svc.Submit(context.Background(), validRequest(), strings.Repeat("1", 80), longAgent); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	stored := repo.created[0]
	if len(stored.IPAddress) > 45 {
		t.Errorf("ip_address would overflow its column: %d bytes", len(stored.IPAddress))
	}
	if len(stored.UserAgent) > 512 {
		t.Errorf("user_agent would overflow its column: %d bytes", len(stored.UserAgent))
	}
}
