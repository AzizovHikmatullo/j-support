package tickets

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

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
 		INSERT INTO tickets(id, category_id, contact_id, status, source) 
 		VALUES ($1, $2, $3, $4, $5) 
 		RETURNING created_at, updated_at
	`

	err := tx.QueryRowxContext(ctx, query,
		ticket.ID,
		ticket.CategoryID,
		ticket.ContactID,
		ticket.Status,
		ticket.Source,
	).Scan(&ticket.CreatedAt, &ticket.UpdatedAt)

	return err
}

func (r *repository) GetByContact(ctx context.Context, creatorID int) ([]Ticket, error) {
	tickets := make([]Ticket, 0)

	query := `
		SELECT * 
		FROM tickets 
		WHERE contact_id = $1 
		ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &tickets, query, creatorID)

	return tickets, err
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

	return tickets, err
}

func (r *repository) GetAll(ctx context.Context) ([]Ticket, error) {
	tickets := make([]Ticket, 0)

	query := `
		SELECT *
		FROM tickets 
		ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &tickets, query)

	return tickets, err
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
	if errors.Is(err, sql.ErrNoRows) {
		return ticket, ErrTicketNotFound
	}

	return ticket, err
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
	if errors.Is(err, sql.ErrNoRows) {
		return ticket, ErrTicketNotFound
	}

	return ticket, err
}

func (r *repository) ChangeStatus(ctx context.Context, status string, ticketID uuid.UUID) error {
	query := `
		UPDATE tickets 
		SET status = $2, updated_at = now() 
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, ticketID, status)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTicketNotFound
	}

	return err
}

func (r *repository) CreateRating(ctx context.Context, rating *Rating) error {
	query := `
        INSERT INTO ticket_ratings(ticket_id, contact_id, score, reason)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `

	err := r.db.QueryRowxContext(ctx, query,
		rating.TicketID,
		rating.ContactID,
		rating.Score,
		rating.Reason,
	).Scan(&rating.ID)

	return err
}

func (r *repository) GetRating(ctx context.Context, ticketID uuid.UUID) (Rating, error) {
	var rating Rating

	query := `
        SELECT id, ticket_id, contact_id, score, reason
        FROM ticket_ratings
        WHERE ticket_id = $1
    `

	err := r.db.GetContext(ctx, &rating, query, ticketID)
	if errors.Is(err, sql.ErrNoRows) {
		return rating, ErrRatingNotFound
	}

	return rating, err
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

	return err
}

func (r *repository) GetMessages(ctx context.Context, ticketID uuid.UUID, limit int, cursor *uuid.UUID) ([]Message, error) {
	messages := make([]Message, 0)

	query := `
        SELECT *
        FROM messages
        WHERE ticket_id = $1
    `

	args := []any{ticketID}

	if cursor != nil {
		query += ` AND id < $2 `
		args = append(args, *cursor)
	}

	query += `
        ORDER BY created_at DESC
        LIMIT $` + strconv.Itoa(len(args)+1)

	args = append(args, limit)

	err := r.db.SelectContext(ctx, &messages, query, args...)

	return messages, err
}

func (r *repository) BeginTxx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}
