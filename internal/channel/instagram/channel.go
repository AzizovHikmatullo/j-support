package instagram

import (
	"context"

	"github.com/AzizovHikmatullo/j-support/internal/channel"
	"github.com/AzizovHikmatullo/j-support/internal/contacts"
)

type instagramChannel struct {
	contactService contacts.Service
}

func New(contactService contacts.Service) channel.Channel {
	return &instagramChannel{contactService: contactService}
}

func (c *instagramChannel) Name() string {
	return channel.ChannelInstagram
}

func (c *instagramChannel) ResolveContact(ctx context.Context, rawID string) (*contacts.Contact, error) {
	contact, err := c.contactService.Resolve(ctx, nil, &rawID, channel.ChannelInstagram)
	if err != nil {
		return nil, err
	}

	return &contact, nil
}
