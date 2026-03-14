package bot

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type postgresRepo struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &postgresRepo{
		db: db,
	}
}

func (r *postgresRepo) GetActiveScenario(ctx context.Context, categoryID int) (Scenario, error) {
	var scenario Scenario

	query := `
		SELECT *
		FROM bot_scenarios
		WHERE category_id = $1 AND is_active = true
	`

	err := r.db.GetContext(ctx, &scenario, query, categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Scenario{}, ErrScenarioNotFound
		}
		return Scenario{}, ErrGetScenario
	}

	return scenario, nil
}

func (r *postgresRepo) GetSteps(ctx context.Context, scenarioID int) ([]Step, error) {
	var steps []Step

	query := `
		SELECT *
		FROM bot_steps
		WHERE scenario_id = $1
		ORDER BY step_order
	`

	err := r.db.SelectContext(ctx, &steps, query, scenarioID)
	if err != nil {
		return nil, ErrGetSteps
	}

	return steps, nil
}

func (r *postgresRepo) CreateSession(ctx context.Context, ticketID uuid.UUID, scenarioID int) error {
	query := `
		INSERT INTO bot_sessions(ticket_id, scenario_id, current_step)
		VALUES ($1, $2, 0)
	`

	err := r.db.QueryRowxContext(ctx, query, ticketID, scenarioID).Err()
	if err != nil {
		return ErrCreateSession
	}

	return nil
}

func (r *postgresRepo) GetSession(ctx context.Context, ticketID uuid.UUID) (Session, error) {
	var session Session

	query := `
		SELECT * 
		FROM bot_sessions 
		WHERE ticket_id = $1
	`

	err := r.db.GetContext(ctx, &session, query, ticketID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, ErrSessionNotFound
		}
		return Session{}, ErrGetSession
	}

	return session, nil
}

func (r *postgresRepo) UpdateSessionStep(ctx context.Context, ticketID uuid.UUID, nextStep int) (Session, error) {
	var session Session

	query := `
		UPDATE bot_sessions 
		SET current_step = $1 
		WHERE ticket_id = $2
		RETURNING *
	`

	err := r.db.QueryRowxContext(ctx, query, nextStep, ticketID).StructScan(&session)
	if err != nil {
		return Session{}, ErrUpdateSession
	}

	return session, nil
}
