package tickets

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
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

func (r *repository) Create(ctx context.Context, tx *sqlx.Tx, ticket *Ticket) error {
	query := `
 		INSERT INTO tickets(id, category_id, creator_id, status, subject, source) 
 		VALUES ($1, $2, $3, $4, $5, $6) 
 		RETURNING created_at, updated_at
	`

	err := tx.QueryRowxContext(ctx, query,
		ticket.ID,
		ticket.CategoryID,
		ticket.CreatorID,
		ticket.Status,
		ticket.Subject,
		ticket.Source,
	).Scan(&ticket.CreatedAt, &ticket.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) GetByCreator(ctx context.Context, creatorID int) ([]Ticket, error) {
	tickets := make([]Ticket, 0)

	query := `
		SELECT * 
		FROM tickets 
		WHERE creator_id = $1 
		ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &tickets, query, creatorID)
	if err != nil {
		return tickets, err
	}

	return tickets, nil
}

func (r *repository) GetSupportTickets(ctx context.Context, assignedTo int) ([]Ticket, error) {
	tickets := make([]Ticket, 0)

	query := `
		SELECT *
		FROM tickets 
		WHERE status = $1 OR assigned_id = $2 
		ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &tickets, query,
		statusOpen,
		assignedTo,
	)
	if err != nil {
		return tickets, err
	}

	return tickets, nil
}

func (r *repository) GetAll(ctx context.Context) ([]Ticket, error) {
	tickets := make([]Ticket, 0)

	query := `
		SELECT *
		FROM tickets 
		ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &tickets, query)
	if err != nil {
		return tickets, err
	}

	return tickets, nil
}

func (r *repository) GetByID(ctx context.Context, ticketID uuid.UUID) (Ticket, error) {
	var ticket Ticket

	query := `
		SELECT *
		FROM tickets 
		WHERE id = $1 
		ORDER BY created_at DESC
	`

	err := r.db.GetContext(ctx, &ticket, query, ticketID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ticket, ErrTicketNotFound
		}
		return ticket, err
	}

	return ticket, nil
}

func (r *repository) ChangeAssigned(ctx context.Context, ticketID uuid.UUID, assignedTo int) (Ticket, error) {
	var ticket Ticket

	query := `
		UPDATE tickets 
		SET assigned_id = $2, status = $3, updated_at = now() 
		WHERE id = $1 
		RETURNING *
	`

	err := r.db.QueryRowxContext(ctx, query,
		ticketID,
		assignedTo,
		statusInProgress,
	).StructScan(&ticket)
	if err != nil {
		return ticket, err
	}

	return ticket, nil
}

func (r *repository) ChangeStatus(ctx context.Context, status string, ticketID uuid.UUID) (Ticket, error) {
	var ticket Ticket

	query := `
		UPDATE tickets 
		SET status = $2, updated_at = now() 
		WHERE id = $1 
		RETURNING *
	`

	err := r.db.QueryRowxContext(ctx, query,
		ticketID,
		status,
	).StructScan(&ticket)
	if err != nil {
		return ticket, err
	}

	return ticket, nil
}

func (r *repository) CreateMessage(ctx context.Context, tx *sqlx.Tx, message *Message) error {
	query := `
		INSERT INTO messages(id, ticket_id, sender_id, sender_type, content) 
		SELECT $1, $2, $3, $4, $5 
		FROM tickets 
		WHERE id = $2 AND status != 'closed' 
		RETURNING created_at
	`

	err := tx.QueryRowxContext(ctx, query,
		message.ID,
		message.TicketID,
		message.SenderID,
		message.SenderType,
		message.Content,
	).StructScan(message)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrClosedTicket
		}
		return err
	}

	return nil
}

func (r *repository) GetMessages(ctx context.Context, ticketID uuid.UUID) ([]Message, error) {
	messages := make([]Message, 0)

	query := `
		SELECT * 
		FROM messages 
		WHERE ticket_id = $1 
		ORDER BY created_at
	`

	err := r.db.SelectContext(ctx, &messages, query, ticketID)
	if err != nil {
		return messages, err
	}

	return messages, nil
}

func (r *repository) BeginTxx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}
