package channel

import (
	"context"

	"github.com/AzizovHikmatullo/j-support/internal/contacts"
)

type Channel interface {
	ResolveContact(ctx context.Context, ID string) (*contacts.Contact, error)
	Name() string
}
