package categories

import (
	"context"
	"fmt"
)

type Repository interface {
	Create(ctx context.Context, name, destination string) (Category, error)
	GetAll(ctx context.Context) ([]Category, error)
	GetForDest(ctx context.Context, destination string) ([]Category, error)
	Update(ctx context.Context, id int, name *string, enabled *bool) (Category, error)
	GetByID(ctx context.Context, id int) (Category, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) Create(ctx context.Context, name, destination string) (Category, error) {
	if name == "" {
		return Category{}, ErrInvalidName
	}

	dest, ok := destinationMapping[destination]
	if !ok {
		return Category{}, ErrInvalidDest
	}

	category, err := s.repo.Create(ctx, name, dest)
	if err != nil {
		return Category{}, fmt.Errorf("create category: %w", err)
	}
	return category, nil
}

func (s *service) Get(ctx context.Context, role string) ([]Category, error) {
	if role == "admin" || role == "support" {
		categories, err := s.repo.GetAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("get category for admins: %w", err)
		}
		return categories, nil
	}

	categories, err := s.repo.GetForDest(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("get category for dest: %w", err)
	}
	return categories, nil
}

func (s *service) Update(ctx context.Context, id int, name *string, enabled *bool) (Category, error) {
	if name != nil && *name == "" {
		return Category{}, ErrInvalidName
	}

	updatedCategory, err := s.repo.Update(ctx, id, name, enabled)
	if err != nil {
		return Category{}, fmt.Errorf("update category: %w", err)
	}
	return updatedCategory, nil
}
