package scenario

import (
	"context"
	"database/sql"
	"errors"

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
	if err != nil {
		return Scenario{}, ErrCreateScenario
	}

	return scenario, nil
}

func (r *postgresRepo) CreateSteps(ctx context.Context, scenarioID int, req []StepRequest) ([]Step, error) {
	steps := make([]Step, 0, len(req))
	questions := make([]string, 0, len(req))

	builder := squirrel.Insert("bot_steps").
		Columns("scenario_id", "step_order", "question").
		PlaceholderFormat(squirrel.Dollar).
		Suffix("RETURNING id")

	for i, step := range req {
		builder = builder.Values(scenarioID, i+1, step.Question)
		questions = append(questions, step.Question)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	var ids []int
	err = r.db.SelectContext(ctx, &ids, query, args...)
	if err != nil {
		return nil, ErrCreateSteps
	}

	for i, id := range ids {
		steps = append(steps, Step{
			ID:         id,
			ScenarioID: scenarioID,
			StepOrder:  i + 1,
			Question:   questions[i],
		})
	}

	return steps, nil
}

func (r *postgresRepo) Get(ctx context.Context, id int) (Scenario, error) {
	var scenario Scenario

	query := `
		SELECT *
		FROM bot_scenarios
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &scenario, query, id)
	if err != nil {
		return Scenario{}, ErrGetScenario
	}

	return scenario, nil
}

func (r *postgresRepo) GetAll(ctx context.Context) ([]Scenario, error) {
	var scenarios []Scenario

	query := `
		SELECT *
		FROM bot_scenarios
	`

	err := r.db.SelectContext(ctx, &scenarios, query)
	if err != nil {
		return nil, ErrGetScenario
	}

	return scenarios, nil
}

func (r *postgresRepo) Update(ctx context.Context, scenarioID int, req UpdateScenarioRequest) (Scenario, error) {
	tx, err := r.db.BeginTxx(ctx, nil)

	updateBuilder := squirrel.Update("bot_scenarios").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": scenarioID})

	if req.CategoryID != nil {
		updateBuilder = updateBuilder.Set("category_id", *req.CategoryID)
	}

	if req.IsActive != nil {
		updateBuilder = updateBuilder.Set("is_active", *req.IsActive)
	}

	query, args, err := updateBuilder.ToSql()
	if err != nil {
		return Scenario{}, err
	}

	if len(args) > 0 {
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return Scenario{}, err
		}
	}

	if req.Steps != nil {
		_, err := tx.ExecContext(ctx,
			`DELETE FROM bot_steps WHERE scenario_id=$1`,
			scenarioID,
		)
		if err != nil {
			return Scenario{}, err
		}

		if len(*req.Steps) > 0 {
			insertBuilder := squirrel.Insert("bot_steps").
				Columns("scenario_id", "step_order", "question").
				PlaceholderFormat(squirrel.Dollar)

			for i, step := range *req.Steps {
				insertBuilder = insertBuilder.Values(scenarioID, i+1, step.Question)
			}

			query, args, err := insertBuilder.ToSql()
			if err != nil {
				return Scenario{}, err
			}

			_, err = tx.ExecContext(ctx, query, args...)
			if err != nil {
				return Scenario{}, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return Scenario{}, err
	}

	return r.Get(ctx, scenarioID)
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
