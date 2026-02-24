package categories

import (
	"errors"
	"time"
)

var (
	ErrForbidden        = errors.New("forbidden")
	ErrInvalidName      = errors.New("invalid name")
	ErrInvalidDest      = errors.New("invalid dest")
	ErrCategoryNotFound = errors.New("category not found")
)

type Category struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Enabled     bool      `json:"enabled" db:"enabled"`
	Destination string    `json:"destination" db:"destination"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"-" db:"updated_at"`
}

type CreateCategoryRequest struct {
	Name        string `json:"name"`
	Destination string `json:"destination"`
}

type UpdateCategoryRequest struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}
