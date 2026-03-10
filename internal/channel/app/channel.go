package app

import (
	"context"

	"github.com/AzizovHikmatullo/j-support/internal/channel"
	"github.com/AzizovHikmatullo/j-support/internal/contacts"
)

type appChannel struct {
	contactService contacts.Service
}

func New(contactService contacts.Service) channel.Channel {
	return &appChannel{contactService: contactService}
}

func (c *appChannel) Name() string {
	return channel.ChannelApp
}

func (c *appChannel) ResolveContact(ctx context.Context, rawID string) (*contacts.Contact, error) {
	contact, err := c.contactService.Resolve(ctx, &rawID, nil)
	if err != nil {
		return nil, err
	}

	return &contact, nil
}
