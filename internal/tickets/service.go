package tickets

import (
	"context"
	"errors"
	"github.com/AzizovHikmatullo/j-support/internal/activity_log"
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
	ChangeStatus(ctx context.Context, status string, ticketID uuid.UUID) error

	CreateRating(ctx context.Context, rating *Rating) error
	GetRating(ctx context.Context, ticketID uuid.UUID) (Rating, error)

	CreateMessage(ctx context.Context, tx *sqlx.Tx, message *Message) error
	GetMessages(ctx context.Context, ticketID uuid.UUID) ([]Message, error)
	BeginTxx(ctx context.Context) (*sqlx.Tx, error)
}

type scenarioService interface {
	StartIfExists(ctx context.Context, ticketID uuid.UUID, categoryID int) (*Message, []string, error)
	HandleMessage(ctx context.Context, ticketID uuid.UUID, answer string) (*string, error)
	GetButtonsForCurrentStep(ctx context.Context, ticketID uuid.UUID) ([]string, error)
}

type service struct {
	repo            Repository
	scenarioService scenarioService
	activityLog     activity_log.Service
	categoryRepo    categories.Repository
	publisher       ws.Publisher
}

func NewService(repo Repository, categoryRepo categories.Repository, pub ws.Publisher, botService scenarioService, al activity_log.Service) Service {
	return &service{
		repo:            repo,
		categoryRepo:    categoryRepo,
		publisher:       pub,
		scenarioService: botService,
		activityLog:     al,
	}
}

func (s *service) SetScenarioService(botService scenarioService) {
	s.scenarioService = botService
}

func (s *service) Create(ctx context.Context, contactID int, role string, source string, req CreateTicketRequest) (*CreateTicketResponse, error) {
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

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	s.activityLog.Log(ctx, activity_log.LogEntry{
		TicketID:  ticket.ID,
		ActorID:   contactID,
		ActorType: role,
		Action:    activity_log.ActionCreated,
		Payload:   activity_log.Payload{"category_id": req.CategoryID, "source": source},
	})

	firstBotMessage, buttons, err := s.scenarioService.StartIfExists(ctx, ticket.ID, category.ID)
	if err != nil {
		return nil, err
	}

	updatedTicket, err := s.repo.GetByID(ctx, ticket.ID)
	if err != nil {
		return nil, err
	}

	return &CreateTicketResponse{
		Ticket: &updatedTicket,
		FirstMessage: &MessageWithButtons{
			Message: firstBotMessage,
			Buttons: buttons,
		},
	}, nil
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

	newTicket, err := s.repo.ChangeAssigned(ctx, ticket.ID, assignedTo)

	s.activityLog.Log(ctx, activity_log.LogEntry{
		TicketID:  ticketID,
		ActorID:   userID,
		ActorType: role,
		Action:    activity_log.ActionAssigned,
		Payload:   activity_log.Payload{"assigned_to": assignedTo},
	})

	event := ws.Event{
		Type:    "assigned_changed",
		Payload: map[string]any{"ticket_id": ticketID, "assigned_to": assignedTo},
	}
	if err = s.publisher.PublishToTicket(ticketID, event); err != nil {
		return Ticket{}, err
	}

	return newTicket, nil
}

func (s *service) ChangeStatus(ctx context.Context, userID int, role string, ticketID uuid.UUID, status string) error {
	if !checkStatus(status) {
		return ErrInvalidStatus
	}

	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return err
	}

	prevStatus := ticket.Status

	if role == "user" && (ticket.ContactID != userID || status != statusClosed) {
		return ErrForbidden
	}

	if role == "support" {
		if ticket.AssignedTo == nil || *ticket.AssignedTo != userID {
			return ErrForbidden
		}
	}

	err = s.repo.ChangeStatus(ctx, status, ticketID)
	if err != nil {
		return err
	}

	s.activityLog.Log(ctx, activity_log.LogEntry{
		TicketID:  ticketID,
		ActorID:   userID,
		ActorType: role,
		Action:    activity_log.ActionStatusChanged,
		Payload:   activity_log.Payload{"from": prevStatus, "to": status},
	})

	event := ws.Event{
		Type:    "status_changed",
		Payload: map[string]any{"ticket_id": ticketID, "status": status},
	}
	if err = s.publisher.PublishToTicket(ticketID, event); err != nil {
		return err
	}

	return nil
}

func (s *service) RateTicket(ctx context.Context, contactID int, ticketID uuid.UUID, req CreateRatingRequest) (Rating, error) {
	if req.Score < 1 || req.Score > 5 {
		return Rating{}, ErrInvalidScore
	}

	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return Rating{}, err
	}

	if ticket.Status != statusClosed {
		return Rating{}, ErrNotClosed
	}

	if ticket.ContactID != contactID {
		return Rating{}, ErrForbidden
	}

	_, err = s.repo.GetRating(ctx, ticketID)
	if err == nil {
		return Rating{}, ErrAlreadyRated
	}
	if !errors.Is(err, ErrNotFound) {
		return Rating{}, err
	}

	rating := &Rating{
		TicketID:  ticketID,
		ContactID: contactID,
		Score:     req.Score,
		Reason:    req.Reason,
	}
	if err = s.repo.CreateRating(ctx, rating); err != nil {
		return Rating{}, err
	}

	s.activityLog.Log(ctx, activity_log.LogEntry{
		TicketID:  ticketID,
		ActorID:   contactID,
		ActorType: activity_log.ActorUser,
		Action:    activity_log.ActionRated,
		Payload:   activity_log.Payload{"score": req.Score, "reason": req.Reason},
	})

	return *rating, nil
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

	message, err := s.saveMessage(ctx, tx, &ticket, senderID, senderType, content)
	if err != nil {
		return nil, err
	}

	publishError := s.publishMessage(ticket.ID, message, nil)
	if publishError != nil {
		return nil, publishError
	}

	if ticket.Status == statusPending && senderType == "user" {
		if err = s.processScenario(ctx, ticket, message); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	s.logMessage(ctx, ticketID, senderID, senderType, message.Content)

	return message, nil
}

func (s *service) CreateMessageWithButtons(ctx context.Context, ticketID uuid.UUID, senderID int, senderType, content string, buttons []string) (*Message, error) {
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

	message, err := s.saveMessage(ctx, tx, &ticket, senderID, senderType, content)
	if err != nil {
		return nil, err
	}

	publishError := s.publishMessage(ticket.ID, message, buttons)
	if publishError != nil {
		return nil, publishError
	}

	if ticket.Status == statusPending && senderType == "user" {
		if err = s.processScenario(ctx, ticket, message); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	s.logMessage(ctx, ticketID, senderID, senderType, message.Content)

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

func (s *service) saveMessage(ctx context.Context, tx *sqlx.Tx, ticket *Ticket, senderID int, senderType string, content string) (*Message, error) {
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

	message := NewMessage(ticket.ID, senderID, senderType, content)

	err := s.repo.CreateMessage(ctx, tx, message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (s *service) publishMessage(ticketID uuid.UUID, message *Message, buttons []string) error {
	event := ws.Event{
		Type: "message_created",
		Payload: map[string]any{
			"message": message,
			"buttons": buttons,
		},
	}

	if err := s.publisher.PublishToTicket(ticketID, event); err != nil {
		return ErrPublishFailed
	}
	return nil
}

func (s *service) processScenario(ctx context.Context, ticket Ticket, message *Message) error {
	nextQuestion, err := s.scenarioService.HandleMessage(ctx, ticket.ID, message.Content)
	if err != nil {
		return nil
	}

	if nextQuestion != nil {
		buttons, err := s.scenarioService.GetButtonsForCurrentStep(ctx, ticket.ID)
		if err != nil {
			return err
		}

		_, err = s.CreateMessageWithButtons(ctx, ticket.ID, 0, "bot", *nextQuestion, buttons)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *service) logMessage(ctx context.Context, ticketID uuid.UUID, actorID int, actorType string, messageContent string) {
	s.activityLog.Log(ctx, activity_log.LogEntry{
		TicketID:  ticketID,
		ActorID:   actorID,
		ActorType: actorType,
		Action:    activity_log.ActionMessageSent,
		Payload:   activity_log.Payload{"message": messageContent},
	})
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
		if ticket.Status == statusOpen || ticket.Status == statusPending {
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

func checkStatus(status string) bool {
	return status == statusOpen || status == statusInProgress || status == statusClosed || status == statusPending
}
