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
	ID        int       `json:"id" db:"id"`
	TicketID  uuid.UUID `json:"ticket_id" db:"ticket_id"`
	ContactID int       `json:"contact_id" db:"contact_id"`
	Score     int       `json:"score" db:"score"`
	Reason    *string   `json:"reason,omitempty" db:"reason"`
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
	Content string `json:"content" binding:"required,min=1,max=150"`
}

type CreateRatingRequest struct {
	Score  int     `json:"score" binding:"required"`
	Reason *string `json:"reason"`
}

var (
	ErrForbidden          = errors.New("forbidden")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrRatingNotFound     = errors.New("rating not found")
	ErrTicketNotFound     = errors.New("ticket not found")
	ErrUnknownChannel     = errors.New("unknown request channel")
	ErrInvalidStatus      = errors.New("invalid status")
	ErrClosedTicket       = errors.New("cannot write to closed ticket")
	ErrNotClosed          = errors.New("ticket is not closed yet")
	ErrCategoryDisabled   = errors.New("category disabled")
	ErrCannotAssign       = errors.New("you can not assign this ticket")
	ErrSupportCannotWrite = errors.New("you cannot write to this ticket")
	ErrAlreadyRated       = errors.New("ticket already rated")
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
