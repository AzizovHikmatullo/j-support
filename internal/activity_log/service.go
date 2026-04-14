package activity_log

import (
	"context"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, entry LogEntry) error
	GetAll(ctx context.Context) ([]ActivityLog, error)
	GetByTicket(ctx context.Context, ticketID uuid.UUID) ([]ActivityLog, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) Log(ctx context.Context, entry LogEntry) {
	_ = s.repo.Create(ctx, entry)
}

func (s *service) GetAll(ctx context.Context) ([]ActivityLog, error) {
	return s.repo.GetAll(ctx)
}

func (s *service) GetByTicket(ctx context.Context, ticketID uuid.UUID) ([]ActivityLog, error) {
	return s.repo.GetByTicket(ctx, ticketID)
}
