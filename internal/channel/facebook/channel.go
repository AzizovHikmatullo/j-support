package facebook

import (
	"context"

	"github.com/AzizovHikmatullo/j-support/internal/channel"
	"github.com/AzizovHikmatullo/j-support/internal/contacts"
)

type facebookChannel struct {
	contactService contacts.Service
}

func New(contactService contacts.Service) channel.Channel {
	return &facebookChannel{contactService: contactService}
}

func (c *facebookChannel) Name() string {
	return channel.ChannelFacebook
}

func (c *facebookChannel) ResolveContact(ctx context.Context, rawID string) (*contacts.Contact, error) {
	contact, err := c.contactService.Resolve(ctx, nil, &rawID, channel.ChannelFacebook)
	if err != nil {
		return nil, err
	}

	return &contact, nil
}
