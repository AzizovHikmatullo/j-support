package ws

import (
	"fmt"

	"github.com/google/uuid"
)

type Publisher interface {
	PublishToTicket(ticketID uuid.UUID, event Event) error
}

type WebSocketPublisher struct {
	hub *Hub
}

func NewPublisher(hub *Hub) *WebSocketPublisher {
	return &WebSocketPublisher{hub: hub}
}

func (p *WebSocketPublisher) PublishToTicket(ticketID uuid.UUID, event Event) error {
	room := fmt.Sprintf("ticket:%s", ticketID.String())
	return p.hub.Broadcast(room, event)
}
