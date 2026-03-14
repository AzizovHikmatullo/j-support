package scenario

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

type Scenario struct {
	ID         int       `json:"id" db:"id"`
	CategoryID int       `json:"category_id" db:"category_id"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	BotSteps   []Step    `json:"-" db:"-"`
}

type Step struct {
	ID         int    `json:"id" db:"id"`
	ScenarioID int    `json:"scenario_id" db:"scenario_id"`
	StepOrder  int    `json:"step_order" db:"step_order"`
	Question   string `json:"question" db:"question"`
	Field      string `json:"field" db:"field"`
}

type Session struct {
	TicketID    uuid.UUID `json:"ticket_id" db:"ticket_id"`
	ScenarioID  int       `json:"scenario_id" db:"scenario_id"`
	CurrentStep int       `json:"current_step" db:"current_step"`
}

var (
	ErrGetScenario      = errors.New("failed to get scenario")
	ErrGetSteps         = errors.New("failed to get steps for scenario")
	ErrCreateSession    = errors.New("failed to create session")
	ErrGetSession       = errors.New("failed to get session by ticket id")
	ErrUpdateSession    = errors.New("failed to update session")
	ErrScenarioNotFound = errors.New("scenario not found for this category")
	ErrSessionNotFound  = errors.New("session not found for this ticket")
)
