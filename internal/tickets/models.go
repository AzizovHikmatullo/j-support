package tickets

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Ticket struct {
	ID         uuid.UUID `json:"id" db:"id"`
	CategoryID int       `json:"category_id" db:"category_id"`
	CreatorID  int       `json:"creator_id" db:"creator_id"`
	AssignedTo *int      `json:"assigned_to" db:"assigned_id"`
	Status     string    `json:"status" db:"status"`
	Subject    string    `json:"subject" db:"subject"`
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

type CreateTicketRequest struct {
	CategoryID int    `json:"category_id" binding:"required"`
	Message    string `json:"message" binding:"required"`
	Subject    string `json:"subject" binding:"required"`
	Source     string `json:"source" binding:"required"`
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

var (
	ErrUndefined          = errors.New("something went wrong")
	ErrCategoryDisabled   = errors.New("category disabled")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidStatus      = errors.New("invalid status")
	ErrInvalidSource      = errors.New("invalid source")
	ErrCannotAssign       = errors.New("you can not assign this ticket")
	ErrTicketNotFound     = errors.New("ticket not found")
	ErrClosedTicket       = errors.New("cannot write to closed ticket")
	ErrSupportCannotWrite = errors.New("you cannot write to this ticket")
	ErrPublishFailed      = errors.New("failed to publish message")
)

const (
	statusOpen       = "open"
	statusInProgress = "in_progress"
	statusClosed     = "closed"

	sourceWeb     = "web"
	sourceMobile  = "mobile"
	sourceService = "service"
)

func NewTicket(creatorID int, req CreateTicketRequest) *Ticket {
	return &Ticket{
		ID:         uuid.Must(uuid.NewV7()),
		CategoryID: req.CategoryID,
		CreatorID:  creatorID,
		Status:     statusOpen,
		Subject:    req.Subject,
		Source:     req.Source,
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
