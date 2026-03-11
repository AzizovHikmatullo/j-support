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

type InitWebRequest struct {
	Name  string `json:"name"  binding:"required"`
	Phone string `json:"phone" binding:"required"`
}

type InitTelegramRequest struct {
	TelegramID string `json:"telegram_id"  binding:"required"`
	Name       string `json:"name"  binding:"required"`
	Phone      string `json:"phone" binding:"required"`
}
