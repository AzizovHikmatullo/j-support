package categories

import (
	"context"
	"github.com/Masterminds/squirrel"
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
		Name:        name,
		Enabled:     true,
		Destination: destination,
	}

	err := r.db.QueryRowxContext(ctx, "INSERT INTO categories(name, destination) VALUES ($1, $2) RETURNING id, created_at", name, destination).StructScan(&category)

	return category, err
}

func (r *postgresRepo) GetAll(ctx context.Context) ([]Category, error) {
	categories := make([]Category, 0)

	err := r.db.SelectContext(ctx, &categories, "SELECT id, name, enabled, destination, created_at FROM categories ORDER BY id")

	return categories, err
}

func (r *postgresRepo) GetForDest(ctx context.Context, destination string) ([]Category, error) {
	categories := make([]Category, 0)

	err := r.db.SelectContext(ctx, &categories, "SELECT id, name, enabled, destination, created_at FROM categories WHERE destination = $1 AND enabled = true", destination)

	return categories, err
}

func (r *postgresRepo) Update(ctx context.Context, id int, name *string, enabled *bool) (Category, error) {
	var category Category

	builder := squirrel.Update("categories").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id}).
		Set("updated_at", squirrel.Expr("NOW()"))

	if name != nil {
		builder = builder.Set("name", *name)
	}

	if enabled != nil {
		builder = builder.Set("enabled", *enabled)
	}

	builder = builder.Suffix("RETURNING id, name, enabled, destination, created_at")

	query, args, err := builder.ToSql()
	if err != nil {
		return Category{}, err
	}

	err = r.db.QueryRowxContext(ctx, query, args...).StructScan(&category)

	return category, err
}

func (r *postgresRepo) GetByID(ctx context.Context, id int) (Category, error) {
	var category Category

	err := r.db.GetContext(ctx, &category, "SELECT id, name, enabled, destination, created_at FROM categories WHERE id = $1", id)

	return category, err
}
