package userapp

import (
	"context"

	"github.com/AzizovHikmatullo/j-support/internal/channel"
	"github.com/AzizovHikmatullo/j-support/internal/contacts"
)

type driverAppChannel struct {
	contactService contacts.Service
}

func New(contactService contacts.Service) channel.Channel {
	return &driverAppChannel{contactService: contactService}
}

func (c *driverAppChannel) Name() string {
	return channel.ChannelDriverApp
}

func (c *driverAppChannel) ResolveContact(ctx context.Context, rawID string) (*contacts.Contact, error) {
	contact, err := c.contactService.Resolve(ctx, &rawID, nil, channel.ChannelDriverApp)
	if err != nil {
		return nil, err
	}

	return &contact, nil
}
