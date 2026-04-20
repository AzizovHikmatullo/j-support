package categories

import (
	"context"
	"fmt"
	"log/slog"
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

	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
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
	s.logger.Info("category created", "id", category.ID, "name", category.Name)
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
	s.logger.Info("category updated", "id", updatedCategory.ID)
	return updatedCategory, nil
}
