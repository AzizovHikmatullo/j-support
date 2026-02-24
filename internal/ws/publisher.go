package ws

import "fmt"

type Publisher interface {
	PublishToTicket(ticketID int, event Event) error
}

type WebSocketPublisher struct {
	hub *Hub
}

func NewPublisher(hub *Hub) *WebSocketPublisher {
	return &WebSocketPublisher{hub: hub}
}

func (p *WebSocketPublisher) PublishToTicket(ticketID int, event Event) error {
	room := fmt.Sprintf("ticket:%d", ticketID)
	return p.hub.Broadcast(room, event)
}
