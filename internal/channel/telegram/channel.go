package telegram

import (
	"context"

	"github.com/AzizovHikmatullo/j-support/internal/channel"
	"github.com/AzizovHikmatullo/j-support/internal/contacts"
)

type tgChannel struct {
	contactService contacts.Service
}

func New(contactService contacts.Service) channel.Channel {
	return &tgChannel{contactService: contactService}
}

func (c *tgChannel) Name() string {
	return channel.ChannelTelegram
}

func (c *tgChannel) ResolveContact(ctx context.Context, rawID string) (*contacts.Contact, error) {
	contact, err := c.contactService.Resolve(ctx, nil, &rawID)
	if err != nil {
		return nil, err
	}

	return &contact, nil
}
