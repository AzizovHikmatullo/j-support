package categories

import (
	"context"
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

func (r *postgresRepo) Create(ctx context.Context, name, destination string) (Category, error) {
	category := Category{
		Name:    name,
		Enabled: true,
	}

	if err := r.db.QueryRowxContext(ctx, "INSERT INTO categories(name, destination) VALUES ($1, $2) RETURNING id, destination, created_at", name, destination).StructScan(&category); err != nil {
		return Category{}, err
	}

	return category, nil
}

func (r *postgresRepo) GetAll(ctx context.Context) ([]Category, error) {
	categories := make([]Category, 0)

	if err := r.db.SelectContext(ctx, &categories, "SELECT id, name, enabled, destination, created_at FROM categories"); err != nil {
		return categories, err
	}

	return categories, nil
}

func (r *postgresRepo) GetForDest(ctx context.Context, destination string) ([]Category, error) {
	categories := make([]Category, 0)

	if err := r.db.SelectContext(ctx, &categories, "SELECT id, name, enabled, destination, created_at FROM categories WHERE destination = $1 AND enabled = true", destination); err != nil {
		return categories, err
	}

	return categories, nil
}

func (r *postgresRepo) Update(ctx context.Context, id int, name string, enabled bool) (Category, error) {
	var category Category

	if err := r.db.QueryRowxContext(ctx, "UPDATE categories SET name = $2, enabled = $3, updated_at = now() WHERE id = $1 RETURNING id, name, enabled, destination, created_at", id, name, enabled).StructScan(&category); err != nil {
		return Category{}, err
	}

	return category, nil
}
