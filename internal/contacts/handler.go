package contacts

import (
	"context"
)

type Service interface {
	Resolve(ctx context.Context, userID, externalID *string, source string) (Contact, error)
	Update(ctx context.Context, id int, name, phone string) (Contact, error)
	InitContact(ctx context.Context, externalID, name, phone, source string) (Contact, error)
}
