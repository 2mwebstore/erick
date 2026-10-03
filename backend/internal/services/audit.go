package services

import (
	"context"
	"log/slog"

	"github.com/kongchansila/portfolio/backend/internal/models"
	"github.com/kongchansila/portfolio/backend/internal/repositories"
)

// AuditService records who changed what.
type AuditService struct {
	repo *repositories.AuditRepository
	log  *slog.Logger
}

func NewAuditService(repo *repositories.AuditRepository, log *slog.Logger) *AuditService {
	return &AuditService{repo: repo, log: log}
}

// Record never returns an error.
//
// An audit write failing must not fail the operation the user asked for — that
// would turn a logging problem into an outage. It is logged loudly instead, so
// the gap is visible.
func (s *AuditService) Record(ctx context.Context, user *models.User, action, entity, entityID string, detail any, ip string) {
	if s == nil || s.repo == nil {
		return
	}

	entry := &models.AuditEntry{
		Action:    action,
		Entity:    entity,
		EntityID:  entityID,
		Detail:    detail,
		IPAddress: ip,
	}
	if user != nil {
		entry.UserID = &user.ID
		entry.ActorEmail = user.Email
	} else {
		entry.ActorEmail = "system"
	}

	if err := s.repo.Record(ctx, entry); err != nil {
		s.log.Error("audit write failed",
			slog.String("action", action),
			slog.String("entity", entity),
			slog.String("error", err.Error()))
	}
}

func (s *AuditService) Recent(ctx context.Context, limit, offset int) (*repositories.AuditListing, error) {
	return s.repo.Recent(ctx, limit, offset)
}
