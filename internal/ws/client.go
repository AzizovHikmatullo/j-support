package ws

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type Client struct {
	conn      *websocket.Conn
	hub       *Hub
	room      string
	send      chan []byte
	closeOnce sync.Once
}

func NewClient(conn *websocket.Conn, hub *Hub, room string) *Client {
	return &Client{
		conn: conn,
		hub:  hub,
		room: room,
		send: make(chan []byte, 256),
	}

}

func (c *Client) ReadPump() {
	defer c.Close()

	c.conn.SetReadLimit(512)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))

	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.hub.Leave(c.room, c)
		_ = c.conn.Close()
		close(c.send)
	})
}
