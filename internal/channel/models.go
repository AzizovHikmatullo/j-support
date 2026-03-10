package channel

import "errors"

var ErrChannelNotFound = errors.New("channel not found")

const (
	ChannelApp      string = "app"
	ChannelWeb      string = "web"
	ChannelTelegram string = "telegram"
)

type Identity struct {
	ChannelType string
	ID          string
}
