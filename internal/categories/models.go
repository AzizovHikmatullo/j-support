package categories

import (
	"errors"
	"time"
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
	Name    *string `json:"name"`
	Enabled *bool   `json:"enabled"`
}

var destinationMapping = map[string]string{
	"user":   "user",
	"client": "user",
	"driver": "driver",
}

var (
	ErrInvalidName  = errors.New("invalid name")
	ErrInvalidDest  = errors.New("invalid dest")
	ErrUnauthorized = errors.New("unauthorized")
)
