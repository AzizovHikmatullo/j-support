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

func (r *postgresRepo) GetByUserID(ctx context.Context, userID, source string) (Contact, error) {
	var contact Contact

	query := `
		SELECT * 
		FROM contacts 
		WHERE user_id = $1 AND source = $2
	`

	err := r.db.GetContext(ctx, &contact, query, userID, source)
	if errors.Is(err, sql.ErrNoRows) {
		return contact, ErrContactNotFound
	}

	return contact, err
}

func (r *postgresRepo) GetByExternalID(ctx context.Context, externalID, source string) (Contact, error) {
	var contact Contact

	query := `
		SELECT * 
		FROM contacts 
		WHERE external_id = $1 AND source = $2
	`

	err := r.db.GetContext(ctx, &contact, query, externalID, source)
	if errors.Is(err, sql.ErrNoRows) {
		return contact, ErrContactNotFound
	}

	return contact, err
}

func (r *postgresRepo) GetByID(ctx context.Context, id int) (Contact, error) {
	var contact Contact

	query := `
		SELECT * 
		FROM contacts 
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &contact, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return contact, ErrContactNotFound
	}

	return contact, err
}

func (r *postgresRepo) GetByPhone(ctx context.Context, phone, source string) (Contact, error) {
	var contact Contact

	query := `
		SELECT * 
		FROM contacts 
		WHERE phone = $1 AND source = $2
	`

	err := r.db.GetContext(ctx, &contact, query, phone, source)
	if errors.Is(err, sql.ErrNoRows) {
		return contact, ErrContactNotFound
	}

	return contact, err
}

func (r *postgresRepo) Create(ctx context.Context, contact *Contact) error {
	query := `
		INSERT INTO contacts(user_id, external_id, name, phone, source) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, created_at
	`

	err := r.db.QueryRowxContext(ctx, query,
		contact.UserID,
		contact.ExternalID,
		contact.Name,
		contact.Phone,
		contact.Source,
	).Scan(&contact.ID, &contact.CreatedAt)

	return err
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
	if errors.Is(err, sql.ErrNoRows) {
		return contact, ErrContactNotFound
	}

	return contact, err
}
