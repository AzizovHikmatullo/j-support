package activity_log

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, entry LogEntry) error {
	query := `
		INSERT INTO activity_log (ticket_id, actor_id, actor_type, action, payload)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		entry.TicketID,
		entry.ActorID,
		entry.ActorType,
		entry.Action,
		entry.Payload,
	)
	if err != nil {
		return ErrCreate
	}
	return nil
}

func (r *repository) GetAll(ctx context.Context) ([]ActivityLog, error) {
	var logs []ActivityLog
	query := `
		SELECT *
		FROM activity_log
		ORDER BY created_at DESC
	`
	err := r.db.SelectContext(ctx, &logs, query)
	if err != nil {
		return nil, ErrGetAll
	}
	return logs, nil
}

func (r *repository) GetByTicket(ctx context.Context, ticketID uuid.UUID) ([]ActivityLog, error) {
	var logs []ActivityLog
	query := `
		SELECT *
		FROM activity_log
		WHERE ticket_id = $1
		ORDER BY created_at DESC
	`
	err := r.db.SelectContext(ctx, &logs, query, ticketID)
	if err != nil {
		return nil, ErrGetByTicket
	}
	return logs, nil
}
