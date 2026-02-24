package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == "http://localhost:5500"
	},
}

type WSHandler struct {
	hub *Hub
}

func NewWSHandler(hub *Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

func (h *WSHandler) ServeTicketWS(c *gin.Context) {
	ticketIDStr := c.Param("id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	room := "ticket:" + ticketIDStr

	client := NewClient(conn, h.hub, room)
	h.hub.Join(room, client)

	go client.WritePump()
	go client.ReadPump()
}
