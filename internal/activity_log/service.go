package activity_log

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, entry LogEntry) error
	GetAll(ctx context.Context) ([]ActivityLog, error)
	GetByTicket(ctx context.Context, ticketID uuid.UUID) ([]ActivityLog, error)
}

type service struct {
	repo Repository

	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

func (s *service) Log(ctx context.Context, entry LogEntry) {
	err := s.repo.Create(ctx, entry)
	if err != nil {
		s.logger.Error("failed to create log", "error", err.Error())
	}
}

func (s *service) GetAll(ctx context.Context) ([]ActivityLog, error) {
	logs, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get activity logs: %w", err)
	}
	return logs, nil
}

func (s *service) GetByTicket(ctx context.Context, ticketID uuid.UUID) ([]ActivityLog, error) {
	logs, err := s.repo.GetByTicket(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("get activity log by: %w", err)
	}
	return logs, nil
}
