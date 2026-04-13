package scenario

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/AzizovHikmatullo/j-support/internal/tickets"
	"github.com/google/uuid"
)

type Repository interface {
	CreateScenario(ctx context.Context, categoryID int) (Scenario, error)
	GetByID(ctx context.Context, id int) (Scenario, error)
	GetAll(ctx context.Context) ([]Scenario, error)
	Update(ctx context.Context, scenarioID int, req UpdateScenarioRequest) (Scenario, error)
	Delete(ctx context.Context, id int) error
	GetActiveScenario(ctx context.Context, categoryID int) (Scenario, error)

	CreateStep(ctx context.Context, scenarioID int, req CreateStepRequest) (Step, error)
	GetAllSteps(ctx context.Context, scenarioID int) ([]Step, error)
	GetStep(ctx context.Context, stepID int) (Step, error)
	GetRootStep(ctx context.Context, scenarioID int) (Step, error)
	GetChildren(ctx context.Context, parentID int) ([]Step, error)
	UpdateStep(ctx context.Context, stepID int, req UpdateStepRequest) (Step, error)
	DeleteStep(ctx context.Context, stepID int) error

	CreateSession(ctx context.Context, ticketID uuid.UUID, scenarioID, stepID int) error
	GetSession(ctx context.Context, ticketID uuid.UUID) (Session, error)
	GetInactiveSessions(ctx context.Context, cutoff time.Time) ([]Session, error)
	UpdateSession(ctx context.Context, ticketID uuid.UUID, nextStepID int) error
	UpdateLastActivity(ctx context.Context, ticketID uuid.UUID) error
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

func (s *service) CreateScenario(ctx context.Context, req CreateScenarioRequest) (Scenario, error) {
	scenario, err := s.repo.CreateScenario(ctx, req.CategoryID)
	if err != nil {
		return Scenario{}, err
	}

	return scenario, nil
}

func (s *service) GetByID(ctx context.Context, id int) (Scenario, error) {
	scenario, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Scenario{}, err
	}

	steps, err := s.repo.GetAllSteps(ctx, id)
	if err != nil {
		return Scenario{}, err
	}

	scenario.BotSteps = buildTree(steps)

	return scenario, nil
}

func (s *service) GetAll(ctx context.Context) ([]Scenario, error) {
	scenarios, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	for i, sc := range scenarios {
		steps, err := s.repo.GetAllSteps(ctx, sc.ID)
		if err != nil {
			return nil, err
		}
		scenarios[i].BotSteps = buildTree(steps)
	}

	return scenarios, nil
}

func (s *service) Update(ctx context.Context, id int, req UpdateScenarioRequest) (Scenario, error) {
	scenario, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return Scenario{}, err
	}

	steps, err := s.repo.GetAllSteps(ctx, id)
	if err != nil {
		return Scenario{}, err
	}

	scenario.BotSteps = buildTree(steps)
	return scenario, nil
}

func (s *service) Delete(ctx context.Context, id int) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *service) CreateStep(ctx context.Context, scenarioID int, req CreateStepRequest) (Step, error) {
	_, err := s.repo.GetByID(ctx, scenarioID)
	if err != nil {
		return Step{}, err
	}

	if req.ParentID == nil {
		_, err = s.repo.GetRootStep(ctx, scenarioID)
		if err == nil {
			return Step{}, ErrRootAlreadyExists
		}
		if !errors.Is(err, ErrStepNotFound) {
			return Step{}, err
		}
	} else {
		parent, err := s.repo.GetStep(ctx, *req.ParentID)
		if errors.Is(err, ErrStepNotFound) {
			return Step{}, ErrParentNotFound
		}
		if err != nil {
			return Step{}, err
		}
		if parent.ScenarioID != scenarioID {
			return Step{}, ErrWrongScenario
		}

		if req.Condition == nil {
			children, err := s.repo.GetChildren(ctx, *req.ParentID)
			if err != nil {
				return Step{}, err
			}
			for _, ch := range children {
				if ch.Condition == nil {
					return Step{}, ErrDefaultAlreadyExists
				}
			}
		}
	}

	return s.repo.CreateStep(ctx, scenarioID, req)
}

func (s *service) GetButtonsForCurrentStep(ctx context.Context, ticketID uuid.UUID) ([]string, error) {
	session, err := s.repo.GetSession(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	children, err := s.repo.GetChildren(ctx, session.CurrentStepID)
	if err != nil || len(children) == 0 {
		return nil, err
	}

	var buttons []string
	for _, ch := range children {
		if ch.Condition != nil && *ch.Condition != "" {
			buttons = append(buttons, *ch.Condition)
		}
	}
	return buttons, nil
}

func (s *service) UpdateStep(ctx context.Context, scenarioID, stepID int, req UpdateStepRequest) (Step, error) {
	step, err := s.repo.GetStep(ctx, stepID)
	if err != nil {
		return Step{}, err
	}
	if step.ScenarioID != scenarioID {
		return Step{}, ErrWrongScenario
	}

	if req.Condition == nil && step.ParentID != nil {
		children, err := s.repo.GetChildren(ctx, *step.ParentID)
		if err != nil {
			return Step{}, err
		}
		for _, ch := range children {
			if ch.Condition == nil && ch.ID != stepID {
				return Step{}, ErrDefaultAlreadyExists
			}
		}
	}

	return s.repo.UpdateStep(ctx, stepID, req)
}

func (s *service) DeleteStep(ctx context.Context, scenarioID, stepID int) error {
	step, err := s.repo.GetStep(ctx, stepID)
	if err != nil {
		return err
	}
	if step.ScenarioID != scenarioID {
		return ErrWrongScenario
	}
	return s.repo.DeleteStep(ctx, stepID)
}

func (s *service) StartIfExists(ctx context.Context, ticketID uuid.UUID, categoryID int) (*tickets.Message, []string, error) {
	scenario, err := s.repo.GetActiveScenario(ctx, categoryID)
	if errors.Is(err, ErrScenarioNotFound) {
		return nil, nil, s.ticketService.ChangeStatus(ctx, 0, "bot", ticketID, "open")
	}

	if err != nil {
		return nil, nil, err
	}

	rootStep, err := s.repo.GetRootStep(ctx, scenario.ID)
	if errors.Is(err, ErrStepNotFound) {
		return nil, nil, s.ticketService.ChangeStatus(ctx, 0, "bot", ticketID, "open")
	}
	if err != nil {
		return nil, nil, err
	}

	if err := s.repo.CreateSession(ctx, ticketID, scenario.ID, rootStep.ID); err != nil {
		return nil, nil, err
	}

	if err := s.ticketService.ChangeStatus(ctx, 0, "bot", ticketID, "pending"); err != nil {
		return nil, nil, err
	}

	buttons, err := s.GetButtonsForCurrentStep(ctx, ticketID)
	if err != nil {
		return nil, nil, err
	}

	msg, err := s.ticketService.CreateMessageWithButtons(ctx, ticketID, 0, "bot", rootStep.Question, buttons)
	if err != nil {
		return nil, nil, err
	}

	return msg, buttons, nil
}

func (s *service) HandleMessage(ctx context.Context, ticketID uuid.UUID, answer string) (*string, error) {
	session, err := s.repo.GetSession(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if err = s.repo.UpdateLastActivity(ctx, ticketID); err != nil {
		return nil, err
	}

	children, err := s.repo.GetChildren(ctx, session.CurrentStepID)
	if err != nil {
		return nil, err
	}

	next := findNext(children, answer)

	if next == nil {
		return nil, s.ticketService.ChangeStatus(ctx, 0, "bot", ticketID, "open")
	}

	if err := s.repo.UpdateSession(ctx, ticketID, next.ID); err != nil {
		return nil, err
	}

	nextChildren, err := s.repo.GetChildren(ctx, next.ID)
	if err != nil {
		return nil, err
	}

	if len(nextChildren) == 0 {
		if err := s.ticketService.ChangeStatus(ctx, 0, "bot", ticketID, "open"); err != nil {
			return nil, err
		}
	}

	return &next.Question, nil
}

func findNext(children []Step, answer string) *Step {
	var defaultStep *Step
	for i, ch := range children {
		if ch.Condition == nil {
			defaultStep = &children[i]
			continue
		}
		if strings.Contains(strings.ToLower(answer), strings.ToLower(*ch.Condition)) {
			return &children[i]
		}
	}
	return defaultStep
}

func buildTree(steps []Step) []StepNode {
	if len(steps) == 0 {
		return []StepNode{}
	}

	nodes := make(map[int]*StepNode, len(steps))
	for i := range steps {
		nodes[steps[i].ID] = &StepNode{Step: steps[i]}
	}

	var roots []*StepNode

	for i := range steps {
		node := nodes[steps[i].ID]

		if steps[i].ParentID == nil {
			roots = append(roots, node)
			continue
		}

		parent, ok := nodes[*steps[i].ParentID]
		if ok {
			parent.Children = append(parent.Children, node)
		}
	}

	result := make([]StepNode, len(roots))
	for i, r := range roots {
		result[i] = *r
	}

	return result
}
