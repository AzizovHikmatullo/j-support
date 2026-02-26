package tickets

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Create(ctx context.Context, tx *sqlx.Tx, userID, categoryID int, source, subject string) (Ticket, error) {
	ticket := Ticket{
		CategoryID: categoryID,
		CreatorID:  userID,
		Status:     statusOpen,
		Subject:    subject,
		Source:     source,
	}

	if err := tx.QueryRowxContext(ctx, "INSERT INTO tickets(category_id, creator_id, status, subject, source) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at", categoryID, userID, statusOpen, subject, source).StructScan(&ticket); err != nil {
		return Ticket{}, err
	}

	return ticket, nil
}

func (r *repository) GetByCreator(ctx context.Context, creatorID int) ([]Ticket, error) {
	tickets := make([]Ticket, 0)

	if err := r.db.SelectContext(ctx, &tickets, "SELECT id, category_id, creator_id, assigned_id, status, subject, source, created_at, updated_at FROM tickets WHERE creator_id = $1 ORDER BY created_at DESC", creatorID); err != nil {
		return tickets, err
	}

	return tickets, nil
}

func (r *repository) GetSupportTickets(ctx context.Context, assignedTo int) ([]Ticket, error) {
	tickets := make([]Ticket, 0)

	if err := r.db.SelectContext(ctx, &tickets, "SELECT id, category_id, creator_id, assigned_id, status, subject, source, created_at, updated_at FROM tickets WHERE status = $1 OR assigned_id = $2 ORDER BY created_at DESC", statusOpen, assignedTo); err != nil {
		return tickets, err
	}

	return tickets, nil
}

func (r *repository) GetAll(ctx context.Context) ([]Ticket, error) {
	tickets := make([]Ticket, 0)

	if err := r.db.SelectContext(ctx, &tickets, "SELECT id, category_id, creator_id, assigned_id, status, subject, source, created_at, updated_at FROM tickets ORDER BY created_at DESC"); err != nil {
		return tickets, err
	}

	return tickets, nil
}

func (r *repository) GetByID(ctx context.Context, id int) (Ticket, error) {
	var ticket Ticket

	if err := r.db.GetContext(ctx, &ticket, "SELECT id, category_id, creator_id, assigned_id, status, subject, source, created_at, updated_at FROM tickets WHERE id = $1 ORDER BY created_at DESC", id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ticket, ErrTicketNotFound
		}
		return ticket, err
	}

	return ticket, nil
}

func (r *repository) ChangeAssigned(ctx context.Context, ticketID, assignedTo int) (Ticket, error) {
	var ticket Ticket

	if err := r.db.QueryRowxContext(ctx, "UPDATE tickets SET assigned_id = $2, status = $3, updated_at = now() WHERE id = $1 RETURNING id, category_id, creator_id, assigned_id, status, subject, source, created_at, updated_at", ticketID, assignedTo, statusInProgress).StructScan(&ticket); err != nil {
		return ticket, err
	}

	return ticket, nil
}

func (r *repository) ChangeStatus(ctx context.Context, status string, ticketID int) (Ticket, error) {
	var ticket Ticket

	if err := r.db.QueryRowxContext(ctx, "UPDATE tickets SET status = $2, updated_at = now() WHERE id = $1 RETURNING id, category_id, creator_id, assigned_id, status, subject, source, created_at, updated_at", ticketID, status).StructScan(&ticket); err != nil {
		return ticket, err
	}

	return ticket, nil
}

func (r *repository) CreateMessage(ctx context.Context, tx *sqlx.Tx, ticketID, senderID int, senderType, content string) (Message, error) {
	message := Message{
		TicketID:   ticketID,
		SenderID:   senderID,
		SenderType: senderType,
		Content:    content,
	}

	if err := tx.QueryRowxContext(ctx, "INSERT INTO messages(ticket_id, sender_id, sender_type, content) VALUES ($1, $2, $3, $4) RETURNING id, created_at", ticketID, senderID, senderType, content).StructScan(&message); err != nil {
		return Message{}, err
	}

	return message, nil
}

func (r *repository) GetMessages(ctx context.Context, ticketID int) ([]Message, error) {
	messages := make([]Message, 0)

	if err := r.db.SelectContext(ctx, &messages, "SELECT id, ticket_id, sender_id, sender_type, content, created_at FROM messages WHERE ticket_id = $1 ORDER BY created_at", ticketID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return messages, nil
		}
		return nil, err
	}

	return messages, nil
}

func (r *repository) BeginTxx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}
