package scenario

import (
	"context"
	"github.com/google/uuid"
)

type Service interface {
	StartIfExists(ctx context.Context, ticketID uuid.UUID, categoryID int) error
	HandleMessage(ctx context.Context, ticketID uuid.UUID) (*string, error)
}
