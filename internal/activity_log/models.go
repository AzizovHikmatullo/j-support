package activity_log

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type ActivityLog struct {
	ID        int       `json:"id" db:"id"`
	TicketID  uuid.UUID `json:"ticket_id" db:"ticket_id"`
	ActorID   int       `json:"actor_id" db:"actor_id"`
	ActorType string    `json:"actor_type" db:"actor_type"`
	Action    string    `json:"action" db:"action"`
	Payload   Payload   `json:"payload" db:"payload"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Payload map[string]any

func (p Payload) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	b, err := json.Marshal(p)
	return string(b), err
}

func (p *Payload) Scan(src any) error {
	if src == nil {
		*p = nil
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case string:
		b = []byte(v)
	case []byte:
		b = v
	default:
		return errors.New("unsupported type for Payload")
	}
	return json.Unmarshal(b, p)
}

const (
	ActionCreated       = "created"
	ActionStatusChanged = "status_changed"
	ActionAssigned      = "assigned"
	ActionMessageSent   = "message_sent"
	ActionRated         = "rated"

	ActorUser = "user"
)

var (
	ErrGetByTicket = errors.New("failed to get activity log for this ticket")
	ErrGetAll      = errors.New("failed to get full activity log")
	ErrCreate      = errors.New("failed to create log entry")
)

type LogEntry struct {
	TicketID  uuid.UUID
	ActorID   int
	ActorType string
	Action    string
	Payload   Payload
}
