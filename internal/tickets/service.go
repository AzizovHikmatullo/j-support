package tickets

import (
	"context"

	"github.com/AzizovHikmatullo/j-support/internal/categories"
	"github.com/AzizovHikmatullo/j-support/internal/ws"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Create(ctx context.Context, tx *sqlx.Tx, userID, categoryID int, source, subject string) (Ticket, error)
	GetByCreator(ctx context.Context, creatorID int) ([]Ticket, error)
	GetSupportTickets(ctx context.Context, assignedTo int) ([]Ticket, error)
	GetAll(ctx context.Context) ([]Ticket, error)
	GetByID(ctx context.Context, ticketID int) (Ticket, error)
	ChangeAssigned(ctx context.Context, ticketID, assignedTo int) (Ticket, error)
	ChangeStatus(ctx context.Context, status string, ticketID int) (Ticket, error)

	CreateMessage(ctx context.Context, tx *sqlx.Tx, ticketID, senderID int, senderType, content string) (Message, error)
	GetMessages(ctx context.Context, ticketID int) ([]Message, error)
	BeginTxx(ctx context.Context) (*sqlx.Tx, error)
}

type service struct {
	repo         Repository
	categoryRepo categories.Repository
	publisher    ws.Publisher
}

func NewService(repo Repository, categoryRepo categories.Repository, pub ws.Publisher) Service {
	return &service{
		repo:         repo,
		categoryRepo: categoryRepo,
		publisher:    pub,
	}
}

func (s *service) Create(ctx context.Context, userID, categoryID int, role, source, subject, content string) (Ticket, error) {
	tx, err := s.repo.BeginTxx(ctx)
	if err != nil {
		return Ticket{}, ErrUndefined
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	category, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return Ticket{}, err
	}

	if !category.Enabled {
		return Ticket{}, ErrCategoryDisabled
	}

	ticket, err := s.repo.Create(ctx, tx, userID, categoryID, source, subject)
	if err != nil {
		return Ticket{}, err
	}

	_, err = s.repo.CreateMessage(ctx, tx, ticket.ID, userID, role, content)
	if err != nil {
		return Ticket{}, err
	}

	if err := tx.Commit(); err != nil {
		return Ticket{}, err
	}

	return ticket, nil
}

func (s *service) Get(ctx context.Context, role string, userID int) ([]Ticket, error) {
	switch role {
	case "user":
		return s.repo.GetByCreator(ctx, userID)
	case "support":
		return s.repo.GetSupportTickets(ctx, userID)
	case "admin":
		return s.repo.GetAll(ctx)
	default:
		return nil, ErrForbidden
	}
}

func (s *service) GetByID(ctx context.Context, role string, userID, ticketID int) (Ticket, error) {
	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return Ticket{}, err
	}

	messages, err := s.repo.GetMessages(ctx, ticketID)
	if err != nil {
		return Ticket{}, err
	}

	ticket.Messages = messages

	if role == "user" && ticket.CreatorID == userID {
		return ticket, nil
	}

	if role == "support" && ticket.AssignedTo != nil {
		if *ticket.AssignedTo == userID {
			return ticket, nil
		}
	}

	if role == "admin" {
		return ticket, nil
	}

	return Ticket{}, ErrForbidden
}

func (s *service) ChangeAssigned(ctx context.Context, role string, userID, ticketID, assignedTo int) (Ticket, error) {
	if role == "support" && assignedTo != userID {
		return Ticket{}, ErrCannotAssign
	}

	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return Ticket{}, err
	}

	if ticket.Status == statusClosed {
		return Ticket{}, ErrClosedTicket
	}

	return s.repo.ChangeAssigned(ctx, ticket.ID, assignedTo)
}

func (s *service) ChangeStatus(ctx context.Context, role, status string, ticketID, userID int) (Ticket, error) {
	if status != statusOpen && status != statusInProgress && status != statusClosed {
		return Ticket{}, ErrInvalidStatus
	}

	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return Ticket{}, err
	}

	if role == "user" && ticket.CreatorID != userID {
		return Ticket{}, ErrForbidden
	}

	if role == "support" {
		if ticket.AssignedTo == nil || *ticket.AssignedTo != userID {
			return Ticket{}, ErrForbidden
		}
	}

	return s.repo.ChangeStatus(ctx, status, ticketID)
}

func (s *service) CreateMessage(ctx context.Context, ticketID, senderID int, senderType, content string) (Message, error) {
	tx, err := s.repo.BeginTxx(ctx)
	if err != nil {
		return Message{}, ErrUndefined
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return Message{}, err
	}

	if senderType == "user" && ticket.CreatorID != senderID {
		return Message{}, ErrForbidden
	}

	if senderType == "support" && ticket.AssignedTo != nil {
		if *ticket.AssignedTo != senderID {
			return Message{}, ErrForbidden
		}
	}

	message, err := s.repo.CreateMessage(ctx, tx, ticketID, senderID, senderType, content)
	if err != nil {
		return Message{}, err
	}

	event := ws.Event{
		Type:    "message_created",
		Payload: message,
	}

	if err = tx.Commit(); err != nil {
		return Message{}, err
	}

	_ = s.publisher.PublishToTicket(ticket.ID, event)

	return message, nil
}

func (s *service) GetMessages(ctx context.Context, role string, ticketID, userID int) ([]Message, error) {
	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if role == "user" && ticket.CreatorID != userID {
		return nil, ErrForbidden
	}

	messages, err := s.repo.GetMessages(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	return messages, nil
}
