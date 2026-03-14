package bot

import (
	"context"
	"errors"
	"github.com/AzizovHikmatullo/j-support/internal/tickets"
	"github.com/google/uuid"
)

type Repository interface {
	GetActiveScenario(ctx context.Context, categoryID int) (Scenario, error)
	GetSteps(ctx context.Context, scenarioID int) ([]Step, error)
	CreateSession(ctx context.Context, ticketID uuid.UUID, scenarioID int) error
	GetSession(ctx context.Context, ticketID uuid.UUID) (Session, error)
	UpdateSessionStep(ctx context.Context, ticketID uuid.UUID, nextStep int) (Session, error)
}

type service struct {
	repo          Repository
	ticketService tickets.Service
}

func NewService(repo Repository, ticketService tickets.Service) Service {
	return &service{
		repo:          repo,
		ticketService: ticketService,
	}
}

func (s *service) StartIfExists(ctx context.Context, ticketID uuid.UUID, categoryID int) error {
	scenario, err := s.repo.GetActiveScenario(ctx, categoryID)
	if errors.Is(err, ErrScenarioNotFound) {
		s.ticketService.ChangeStatus(ctx, 0, "bot", ticketID, "open")
		return nil
	}

	if err != nil {
		return err
	}

	steps, err := s.repo.GetSteps(ctx, scenario.ID)
	if err != nil {
		return err
	}

	if err := s.repo.CreateSession(ctx, ticketID, scenario.ID); err != nil {
		return err
	}

	if err := s.ticketService.ChangeStatus(ctx, 0, "bot", ticketID, "pending"); err != nil {
		return err
	}

	firstQuestion := steps[0].Question
	_, err = s.ticketService.CreateMessage(ctx, ticketID, 0, "bot", firstQuestion)
	return err
}

func (s *service) HandleMessage(ctx context.Context, ticketID uuid.UUID, answer string) (*string, error) {
	session, err := s.repo.GetSession(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	steps, err := s.repo.GetSteps(ctx, session.ScenarioID)
	if err != nil {
		return nil, err
	}

	nextStepIndex := session.CurrentStep + 1

	if nextStepIndex >= len(steps) {
		if err := s.ticketService.ChangeStatus(ctx, 0, "bot", ticketID, "open"); err != nil {
			return nil, err
		}
		return nil, nil
	}

	if _, err := s.repo.UpdateSessionStep(ctx, ticketID, nextStepIndex); err != nil {
		return nil, err
	}

	nextQuestion := steps[nextStepIndex].Question
	return &nextQuestion, nil
}
