package tickets

import (
	"errors"
	"time"
)

type Ticket struct {
	ID         int       `json:"id" db:"id"`
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
	ID         int64     `json:"id" db:"id"`
	TicketID   int       `json:"ticket_id" db:"ticket_id"`
	SenderID   int       `json:"sender_id" db:"sender_id"`
	SenderType string    `json:"sender_type" db:"sender_type"`
	Content    string    `json:"content" db:"content"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type CreateTicketRequest struct {
	CategoryID int    `json:"category_id"`
	Message    string `json:"message"`
	Subject    string `json:"subject"`
	Source     string `json:"source"`
}

type ChangeAssignedRequest struct {
	AssignedTo int `json:"assigned_to"`
}

type ChangeStatusRequest struct {
	Status string `json:"status"`
}

type CreateMessageRequest struct {
	Content string `json:"content"`
}

var (
	ErrUndefined        = errors.New("something went wrong")
	ErrCategoryDisabled = errors.New("category disabled")
	ErrForbidden        = errors.New("forbidden")
	ErrInvalidStatus    = errors.New("invalid status")
	ErrCannotAssign     = errors.New("you can not assign this ticket")
	ErrTicketNotFound   = errors.New("ticket not found")
	ErrClosedTicket     = errors.New("cannot write to closed ticket")
)

const (
	statusOpen       = "open"
	statusInProgress = "in_progress"
	statusClosed     = "closed"
)
