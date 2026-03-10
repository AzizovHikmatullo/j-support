package contacts

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	Resolve(ctx context.Context, userID, externalID *string) (Contact, error)
	Update(ctx context.Context, id uuid.UUID, name, phone string) (Contact, error)
}
