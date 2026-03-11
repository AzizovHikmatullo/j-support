package tickets

import (
	"context"

	"github.com/AzizovHikmatullo/j-support/internal/categories"
	"github.com/AzizovHikmatullo/j-support/internal/ws"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Create(ctx context.Context, tx *sqlx.Tx, ticket *Ticket) error
	GetByContact(ctx context.Context, creatorID int) ([]Ticket, error)
	GetSupportTickets(ctx context.Context, assignedTo int) ([]Ticket, error)
	GetAll(ctx context.Context) ([]Ticket, error)
	GetByID(ctx context.Context, ticketID uuid.UUID) (Ticket, error)
	ChangeAssigned(ctx context.Context, ticketID uuid.UUID, assignedTo int) (Ticket, error)
	ChangeStatus(ctx context.Context, status string, ticketID uuid.UUID) (Ticket, error)

	CreateMessage(ctx context.Context, tx *sqlx.Tx, message *Message) error
	GetMessages(ctx context.Context, ticketID uuid.UUID) ([]Message, error)
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

func (s *service) Create(ctx context.Context, contactID int, role string, source string, req CreateTicketRequest) (*Ticket, error) {
	tx, err := s.repo.BeginTxx(ctx)
	if err != nil {
		return nil, ErrUndefined
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	category, err := s.categoryRepo.GetByID(ctx, req.CategoryID)
	if err != nil {
		return nil, err
	}

	if !category.Enabled {
		return nil, ErrCategoryDisabled
	}

	ticket := NewTicket(contactID, source, req)

	err = s.repo.Create(ctx, tx, ticket)
	if err != nil {
		return nil, err
	}

	message := NewMessage(ticket.ID, contactID, role, req.Message)

	err = s.repo.CreateMessage(ctx, tx, message)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (s *service) Get(ctx context.Context, role string, userID int) ([]Ticket, error) {
	switch role {
	case "user":
		return s.repo.GetByContact(ctx, userID)
	case "support":
		return s.repo.GetSupportTickets(ctx, userID)
	case "admin":
		return s.repo.GetAll(ctx)
	default:
		return nil, ErrForbidden
	}
}

func (s *service) GetByID(ctx context.Context, userID int, role string, ticketID uuid.UUID) (Ticket, error) {
	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return Ticket{}, err
	}

	if err := checkAccess(userID, role, ticket); err != nil {
		return Ticket{}, err
	}

	messages, err := s.repo.GetMessages(ctx, ticketID)
	if err != nil {
		return Ticket{}, err
	}

	ticket.Messages = messages

	return ticket, nil
}

func (s *service) GetMine(ctx context.Context, contactID int, ticketID uuid.UUID) (Ticket, error) {
	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return Ticket{}, err
	}

	if ticket.ContactID != contactID {
		return Ticket{}, ErrForbidden
	}

	messages, err := s.repo.GetMessages(ctx, ticketID)
	if err != nil {
		return Ticket{}, err
	}

	ticket.Messages = messages

	return ticket, nil
}

func (s *service) ChangeAssigned(ctx context.Context, userID int, role string, ticketID uuid.UUID, assignedTo int) (Ticket, error) {
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

func (s *service) ChangeStatus(ctx context.Context, userID int, role string, ticketID uuid.UUID, status string) (Ticket, error) {
	if status != statusOpen && status != statusInProgress && status != statusClosed {
		return Ticket{}, ErrInvalidStatus
	}

	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return Ticket{}, err
	}

	if role == "user" && (ticket.ContactID != userID || status != statusClosed) {
		return Ticket{}, ErrForbidden
	}

	if role == "support" {
		if ticket.AssignedTo == nil || *ticket.AssignedTo != userID {
			return Ticket{}, ErrForbidden
		}
	}

	return s.repo.ChangeStatus(ctx, status, ticketID)
}

func (s *service) CreateMessage(ctx context.Context, ticketID uuid.UUID, senderID int, senderType, content string) (*Message, error) {
	tx, err := s.repo.BeginTxx(ctx)
	if err != nil {
		return nil, ErrUndefined
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if senderType == "user" && ticket.ContactID != senderID {
		return nil, ErrForbidden
	}

	if senderType == "support" && ticket.AssignedTo == nil {
		return nil, ErrSupportCannotWrite
	}

	if senderType == "support" && ticket.AssignedTo != nil {
		if *ticket.AssignedTo != senderID {
			return nil, ErrForbidden
		}
	}

	message := NewMessage(ticketID, senderID, senderType, content)

	err = s.repo.CreateMessage(ctx, tx, message)
	if err != nil {
		return nil, err
	}

	event := ws.Event{
		Type:    "message_created",
		Payload: message,
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	if err := s.publisher.PublishToTicket(ticket.ID, event); err != nil {
		return nil, ErrPublishFailed
	}

	return message, nil
}

func (s *service) GetMessages(ctx context.Context, userID int, role string, ticketID uuid.UUID) ([]Message, error) {
	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if role == "user" && ticket.ContactID != userID {
		return nil, ErrForbidden
	}

	messages, err := s.repo.GetMessages(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func checkAccess(userID int, role string, ticket Ticket) error {
	switch role {
	case "admin":
		return nil
	case "user":
		if ticket.ContactID != userID {
			return ErrForbidden
		}
		return nil
	case "support":
		if ticket.Status == statusOpen {
			return nil
		}
		if ticket.AssignedTo != nil && *ticket.AssignedTo == userID {
			return nil
		}
		return ErrForbidden
	default:
		return ErrForbidden
	}
}
