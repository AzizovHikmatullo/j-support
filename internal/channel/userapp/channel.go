package userapp

import (
	"context"

	"github.com/AzizovHikmatullo/j-support/internal/channel"
	"github.com/AzizovHikmatullo/j-support/internal/contacts"
)

type userAppChannel struct {
	contactService contacts.Service
}

func New(contactService contacts.Service) channel.Channel {
	return &userAppChannel{contactService: contactService}
}

func (c *userAppChannel) Name() string {
	return channel.ChannelUserApp
}

func (c *userAppChannel) ResolveContact(ctx context.Context, rawID string) (*contacts.Contact, error) {
	contact, err := c.contactService.Resolve(ctx, &rawID, nil)
	if err != nil {
		return nil, err
	}

	return &contact, nil
}
