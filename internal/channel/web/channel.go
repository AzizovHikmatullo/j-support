package web

import (
	"context"

	"github.com/AzizovHikmatullo/j-support/internal/channel"
	"github.com/AzizovHikmatullo/j-support/internal/contacts"
)

type webChannel struct {
	contactService contacts.Service
}

func New(contactService contacts.Service) channel.Channel {
	return &webChannel{contactService: contactService}
}

func (c *webChannel) Name() string {
	return channel.ChannelWeb
}

func (c *webChannel) ResolveContact(ctx context.Context, rawID string) (*contacts.Contact, error) {
	contact, err := c.contactService.Resolve(ctx, nil, &rawID)
	if err != nil {
		return nil, err
	}

	return &contact, nil
}
