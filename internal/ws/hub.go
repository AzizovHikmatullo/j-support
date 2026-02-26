package ws

import (
	"sync"
)

type Hub struct {
	rooms  map[string]map[*Client]struct{}
	mu     sync.RWMutex
	closed bool
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]map[*Client]struct{}),
	}
}

func (h *Hub) Join(room string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return
	}

	if _, ok := h.rooms[room]; !ok {
		h.rooms[room] = make(map[*Client]struct{})
	}
	h.rooms[room][client] = struct{}{}
}

func (h *Hub) Leave(room string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return
	}

	if clients, ok := h.rooms[room]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.rooms, room)
		}
	}
}

func (h *Hub) Broadcast(room string, event Event) error {
	data, err := event.Marshal()
	if err != nil {
		return err
	}

	h.mu.RLock()
	clients, ok := h.rooms[room]
	if !ok {
		h.mu.RUnlock()
		return nil
	}

	var toRemove []*Client
	for client := range clients {
		select {
		case client.send <- data:
		default:
			toRemove = append(toRemove, client)
		}
	}
	h.mu.RUnlock()

	for _, client := range toRemove {
		client.Close()
	}

	return nil
}

func (h *Hub) Shutdown() {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return
	}
	h.closed = true

	var clients []*Client
	for _, room := range h.rooms {
		for client := range room {
			clients = append(clients, client)
		}
	}
	h.rooms = make(map[string]map[*Client]struct{})
	h.mu.Unlock()

	for _, client := range clients {
		client.Close()
	}
}
