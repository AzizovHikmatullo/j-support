package tickets

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Ticket struct {
	ID         uuid.UUID `json:"id" db:"id"`
	CategoryID int       `json:"category_id" db:"category_id"`
	ContactID  int       `json:"creator_id" db:"contact_id"`
	AssignedTo *int      `json:"assigned_to" db:"assigned_id"`
	Status     string    `json:"status" db:"status"`
	Source     string    `json:"source" db:"source"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
	Messages   []Message `json:"messages" db:"-"`
}

type Message struct {
	ID         uuid.UUID `json:"id" db:"id"`
	TicketID   uuid.UUID `json:"ticket_id" db:"ticket_id"`
	SenderID   int       `json:"sender_id" db:"sender_id"`
	SenderType string    `json:"sender_type" db:"sender_type"`
	Content    string    `json:"content" db:"content"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type Rating struct {
	ID        int       `db:"id"         json:"id"`
	TicketID  uuid.UUID `db:"ticket_id"  json:"ticket_id"`
	ContactID int       `db:"contact_id" json:"contact_id"`
	Score     int       `db:"score"      json:"score"`
}

type CreateTicketRequest struct {
	CategoryID int `json:"category_id" binding:"required"`
}

type CreateTicketResponse struct {
	*Ticket
	FirstMessage *MessageWithButtons `json:"first_message,omitempty"`
}

type MessageWithButtons struct {
	*Message
	Buttons []string `json:"buttons,omitempty"`
}

type ChangeAssignedRequest struct {
	AssignedTo int `json:"assigned_to"  binding:"required"`
}

type ChangeStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type CreateMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

type CreateRatingRequest struct {
	Score int `json:"score" binding:"required"`
}

var (
	ErrUndefined          = errors.New("something went wrong")
	ErrCategoryDisabled   = errors.New("category disabled")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidStatus      = errors.New("invalid status")
	ErrCannotAssign       = errors.New("you can not assign this ticket")
	ErrTicketNotFound     = errors.New("ticket not found")
	ErrClosedTicket       = errors.New("cannot write to closed ticket")
	ErrSupportCannotWrite = errors.New("you cannot write to this ticket")
	ErrPublishFailed      = errors.New("failed to publish message")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrUnknownChannel     = errors.New("unknown request channel")
	ErrNotFound           = errors.New("not found")
	ErrAlreadyRated       = errors.New("ticket already rated")
	ErrNotClosed          = errors.New("ticket is not closed yet")
	ErrInvalidScore       = errors.New("score must be between 1 and 5")
)

const (
	statusPending    = "pending"
	statusOpen       = "open"
	statusInProgress = "in_progress"
	statusClosed     = "closed"

	userRole = "user"
)

func NewTicket(contactID int, source string, req CreateTicketRequest) *Ticket {
	return &Ticket{
		ID:         uuid.Must(uuid.NewV7()),
		CategoryID: req.CategoryID,
		ContactID:  contactID,
		Status:     statusPending,
		Source:     source,
	}
}

func NewMessage(ticketID uuid.UUID, senderID int, senderType, content string) *Message {
	return &Message{
		ID:         uuid.Must(uuid.NewV7()),
		TicketID:   ticketID,
		SenderID:   senderID,
		SenderType: senderType,
		Content:    content,
	}
}
