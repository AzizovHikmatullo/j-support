package bot

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	StartIfExists(ctx context.Context, ticketID uuid.UUID, categoryID int) error
	HandleMessage(ctx context.Context, ticketID uuid.UUID, answer string) (*string, error)
}
