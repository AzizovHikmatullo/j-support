package contacts

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrContactNotFound = errors.New("contact not found")
	ErrInvalidPhone    = errors.New("invalid phone")
	ErrInvalidName     = errors.New("invalid name")
	ErrUndefined       = errors.New("undefined")
)

type Contact struct {
	ID         uuid.UUID `db:"id" json:"id"`
	UserID     *string   `db:"user_id" json:"user_id,omitempty"`
	ExternalID *string   `db:"external_id" json:"external_id,omitempty"`
	Name       *string   `db:"name" json:"name,omitempty"`
	Phone      *string   `db:"phone" json:"phone,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type UpdateContactRequest struct {
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone" binding:"required"`
}
