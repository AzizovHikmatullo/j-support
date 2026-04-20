package scenario

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Masterminds/squirrel"
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

func (r *postgresRepo) CreateScenario(ctx context.Context, categoryID int) (Scenario, error) {
	var scenario Scenario

	query := `
		INSERT INTO bot_scenarios(category_id)
		VALUES ($1)
		RETURNING *
	`

	err := r.db.QueryRowxContext(ctx, query, categoryID).StructScan(&scenario)

	return scenario, err
}

func (r *postgresRepo) GetByID(ctx context.Context, id int) (Scenario, error) {
	var scenario Scenario

	query := `
		SELECT *
		FROM bot_scenarios
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &scenario, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return scenario, ErrScenarioNotFound
	}

	return scenario, err
}

func (r *postgresRepo) GetAll(ctx context.Context) ([]Scenario, error) {
	var scenarios []Scenario

	query := `
		SELECT *
		FROM bot_scenarios
		ORDER BY id
	`

	err := r.db.SelectContext(ctx, &scenarios, query)

	return scenarios, err
}

func (r *postgresRepo) Update(ctx context.Context, scenarioID int, req UpdateScenarioRequest) (Scenario, error) {
	var scenario Scenario

	query := `
        UPDATE bot_scenarios
        SET is_active = $2
        WHERE id = $1
        RETURNING *
    `

	err := r.db.QueryRowxContext(ctx, query,
		scenarioID,
		req.IsActive,
	).StructScan(&scenario)
	if errors.Is(err, sql.ErrNoRows) {
		return scenario, ErrScenarioNotFound
	}

	return scenario, err
}

func (r *postgresRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM bot_scenarios WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *postgresRepo) GetActiveScenario(ctx context.Context, categoryID int) (Scenario, error) {
	var scenario Scenario

	query := `
		SELECT *
		FROM bot_scenarios
		WHERE category_id = $1 AND is_active = true
	`

	err := r.db.GetContext(ctx, &scenario, query, categoryID)
	if errors.Is(err, sql.ErrNoRows) {
		return scenario, ErrScenarioNotFound
	}

	return scenario, err
}

func (r *postgresRepo) CreateStep(ctx context.Context, scenarioID int, req CreateStepRequest) (Step, error) {
	var step Step

	query := `
        INSERT INTO bot_steps(scenario_id, parent_id, condition, question)
        VALUES ($1, $2, $3, $4)
        RETURNING *
    `

	err := r.db.QueryRowxContext(ctx, query,
		scenarioID,
		req.ParentID,
		req.Condition,
		req.Question,
	).StructScan(&step)

	return step, err
}

func (r *postgresRepo) GetAllSteps(ctx context.Context, scenarioID int) ([]Step, error) {
	var steps []Step

	query := `
		SELECT *
		FROM bot_steps
		WHERE scenario_id = $1
		ORDER BY id
	`

	err := r.db.SelectContext(ctx, &steps, query, scenarioID)

	return steps, err
}

func (r *postgresRepo) GetStep(ctx context.Context, stepID int) (Step, error) {
	var step Step

	query := `
        SELECT *
        FROM bot_steps
        WHERE id = $1
    `

	err := r.db.GetContext(ctx, &step, query, stepID)
	if errors.Is(err, sql.ErrNoRows) {
		return step, ErrStepNotFound
	}

	return step, err
}

func (r *postgresRepo) GetRootStep(ctx context.Context, scenarioID int) (Step, error) {
	var step Step

	query := `
        SELECT *
        FROM bot_steps
        WHERE scenario_id = $1 AND parent_id IS NULL
    `

	err := r.db.GetContext(ctx, &step, query, scenarioID)
	if errors.Is(err, sql.ErrNoRows) {
		return step, ErrStepNotFound
	}

	return step, err
}

func (r *postgresRepo) GetChildren(ctx context.Context, parentID int) ([]Step, error) {
	var steps []Step

	query := `
        SELECT *
        FROM bot_steps
        WHERE parent_id = $1
    `

	err := r.db.SelectContext(ctx, &steps, query, parentID)

	return steps, err
}

func (r *postgresRepo) UpdateStep(ctx context.Context, stepID int, req UpdateStepRequest) (Step, error) {
	var step Step

	builder := squirrel.Update("bot_steps").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": stepID})

	if req.Condition != nil {
		builder = builder.Set("condition", req.Condition)
	}

	if req.Question != nil {
		builder = builder.Set("question", req.Question)
	}

	builder = builder.Suffix("RETURNING *")

	query, args, err := builder.ToSql()
	if err != nil {
		return Step{}, err
	}

	err = r.db.QueryRowxContext(ctx, query, args...).StructScan(&step)
	if errors.Is(err, sql.ErrNoRows) {
		return step, ErrStepNotFound
	}

	return step, err
}

func (r *postgresRepo) DeleteStep(ctx context.Context, stepID int) error {
	query := `DELETE FROM bot_steps WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, stepID)

	return err
}

func (r *postgresRepo) CreateSession(ctx context.Context, ticketID uuid.UUID, scenarioID, stepID int) error {
	query := `
		INSERT INTO bot_sessions(ticket_id, scenario_id, current_step_id)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(ctx, query, ticketID, scenarioID, stepID)

	return err
}

func (r *postgresRepo) GetSession(ctx context.Context, ticketID uuid.UUID) (Session, error) {
	var session Session

	query := `
		SELECT * 
		FROM bot_sessions 
		WHERE ticket_id = $1
	`

	err := r.db.GetContext(ctx, &session, query, ticketID)
	if errors.Is(err, sql.ErrNoRows) {
		return session, ErrSessionNotFound
	}

	return session, err
}

func (r *postgresRepo) GetInactiveSessions(ctx context.Context, cutoff time.Time) ([]Session, error) {
	query := `
        SELECT bs.ticket_id, bs.scenario_id, bs.current_step_id, bs.created_at, bs.last_activity_at
		FROM bot_sessions bs
		JOIN tickets t ON t.id = bs.ticket_id
		WHERE t.status = 'pending' AND bs.last_activity_at < $1;
    `

	var sessions []Session

	err := r.db.SelectContext(ctx, &sessions, query, cutoff)

	return sessions, err
}

func (r *postgresRepo) UpdateSession(ctx context.Context, ticketID uuid.UUID, nextStepID int) error {
	query := `
        UPDATE bot_sessions
        SET current_step_id = $2
        WHERE ticket_id = $1
    `

	_, err := r.db.ExecContext(ctx, query, ticketID, nextStepID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrSessionNotFound
	}

	return err
}

func (r *postgresRepo) UpdateLastActivity(ctx context.Context, ticketID uuid.UUID) error {
	query := `
        UPDATE bot_sessions
        SET last_activity_at = now()
        WHERE ticket_id = $1
    `

	_, err := r.db.ExecContext(ctx, query, ticketID)

	return err
}
