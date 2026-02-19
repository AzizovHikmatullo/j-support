package categories

import "context"

type Repository interface {
	Create(ctx context.Context, name, destination string) (Category, error)
	GetAll(ctx context.Context) ([]Category, error)
	GetForDest(ctx context.Context, destination string) ([]Category, error)
	Update(ctx context.Context, id int, name string, enabled bool) (Category, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) Create(ctx context.Context, role, name, destination string) (Category, error) {
	if role != "admin" {
		return Category{}, ErrForbidden
	}

	if name == "" {
		return Category{}, ErrInvalidName
	}

	if destination != "user" && destination != "driver" {
		return Category{}, ErrInvalidDest
	}

	return s.repo.Create(ctx, name, destination)
}

func (s *service) Get(ctx context.Context, role string) ([]Category, error) {
	if role == "admin" || role == "support" {
		return s.repo.GetAll(ctx)
	}

	return s.repo.GetForDest(ctx, role)
}

func (s *service) Update(ctx context.Context, role string, id int, name string, enabled bool) (Category, error) {
	if role != "admin" {
		return Category{}, ErrForbidden
	}

	if name == "" {
		return Category{}, ErrInvalidName
	}

	return s.repo.Update(ctx, id, name, enabled)
}
