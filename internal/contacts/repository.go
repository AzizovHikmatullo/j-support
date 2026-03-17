package contacts

import (
	"context"
	"database/sql"
	"errors"
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

func (r *postgresRepo) GetByUserID(ctx context.Context, userID string) (Contact, error) {
	var contact Contact

	query := `
		SELECT * 
		FROM contacts 
		WHERE user_id = $1
	`

	err := r.db.GetContext(ctx, &contact, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return contact, ErrContactNotFound
		}

		return contact, err
	}

	return contact, nil
}

func (r *postgresRepo) GetByExternalID(ctx context.Context, externalID string) (Contact, error) {
	var contact Contact

	query := `
		SELECT * 
		FROM contacts 
		WHERE external_id = $1
	`

	err := r.db.GetContext(ctx, &contact, query, externalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return contact, ErrContactNotFound
		}

		return contact, err
	}

	return contact, nil
}

func (r *postgresRepo) GetByID(ctx context.Context, id int) (Contact, error) {
	var contact Contact

	query := `
		SELECT * 
		FROM contacts 
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &contact, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return contact, ErrContactNotFound
		}

		return contact, err
	}

	return contact, nil
}

func (r *postgresRepo) GetByPhone(ctx context.Context, phone string) (Contact, error) {
	var contact Contact

	query := `
		SELECT * 
		FROM contacts 
		WHERE phone = $1
	`

	err := r.db.GetContext(ctx, &contact, query, phone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return contact, ErrContactNotFound
		}

		return contact, err
	}

	return contact, nil
}

func (r *postgresRepo) Create(ctx context.Context, contact *Contact) error {
	query := `
		INSERT INTO contacts(user_id, external_id, name, phone) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, created_at
	`

	err := r.db.QueryRowxContext(ctx, query,
		contact.UserID,
		contact.ExternalID,
		contact.Name,
		contact.Phone,
	).Scan(&contact.ID, &contact.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *postgresRepo) Update(ctx context.Context, id int, name, phone string) (Contact, error) {
	var contact Contact

	query := `
		UPDATE contacts 
		SET name = $2, phone = $3 
		WHERE id = $1 
		RETURNING *
	`

	err := r.db.QueryRowxContext(ctx, query,
		id,
		name,
		phone,
	).StructScan(&contact)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Contact{}, ErrContactNotFound
		}

		return Contact{}, err
	}

	return contact, nil
}
