package scenario

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Scenario struct {
	ID         int        `json:"id" db:"id"`
	CategoryID int        `json:"category_id" db:"category_id"`
	IsActive   bool       `json:"is_active" db:"is_active"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	BotSteps   []StepNode `json:"steps" db:"-"`
}

type Step struct {
	ID         int       `json:"id" db:"id"`
	ScenarioID int       `json:"scenario_id" db:"scenario_id"`
	ParentID   *int      `json:"parent_id" db:"parent_id"`
	Condition  *string   `json:"condition" db:"condition"`
	Question   string    `json:"question" db:"question"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type Session struct {
	TicketID       uuid.UUID `json:"ticket_id" db:"ticket_id"`
	ScenarioID     int       `json:"scenario_id" db:"scenario_id"`
	CurrentStepID  int       `json:"current_step_id" db:"current_step_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	LastActivityAt time.Time `json:"last_activity_at" db:"last_activity_at"`
}

type StepNode struct {
	Step
	Children []*StepNode `json:"children"`
}

type CreateScenarioRequest struct {
	CategoryID int `json:"category_id" binding:"required"`
}

type UpdateScenarioRequest struct {
	IsActive bool `json:"is_active"`
}

type CreateStepRequest struct {
	ParentID  *int    `json:"parent_id" db:"parent_id"`
	Condition *string `json:"condition" db:"condition"`
	Question  string  `json:"question" binding:"required" db:"question"`
}

type UpdateStepRequest struct {
	Condition *string `json:"condition" db:"condition"`
	Question  *string `json:"question" db:"question"`
}

var (
	ErrGetScenario          = errors.New("failed to get scenario")
	ErrGetSteps             = errors.New("failed to get steps for scenario")
	ErrGetStep              = errors.New("failed to get step by id")
	ErrGetRootStep          = errors.New("failed to get root step by id")
	ErrGetChildren          = errors.New("failed to get root step by id")
	ErrCreateSession        = errors.New("failed to create session")
	ErrGetSession           = errors.New("failed to get session by ticket id")
	ErrUpdateSession        = errors.New("failed to update session")
	ErrUpdateActivity       = errors.New("failed to update session activity")
	ErrUpdateStep           = errors.New("failed to update step")
	ErrCreateScenario       = errors.New("failed to create scenario")
	ErrCreateStep           = errors.New("failed to create step")
	ErrDeleteScenario       = errors.New("failed to delete scenario")
	ErrDeleteStep           = errors.New("failed to delete step")
	ErrScenarioNotFound     = errors.New("scenario not found")
	ErrStepNotFound         = errors.New("step not found")
	ErrSessionNotFound      = errors.New("session not found")
	ErrRootAlreadyExists    = errors.New("scenario already has a root step")
	ErrDefaultAlreadyExists = errors.New("parent already has a default transition")
	ErrParentNotFound       = errors.New("parent step not found")
	ErrWrongScenario        = errors.New("parent step belongs to different scenario")
)
