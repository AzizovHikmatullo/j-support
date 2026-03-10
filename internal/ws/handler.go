package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	room := "ticket:" + ticketID.String()

	client := NewClient(conn, h.hub, room)
	h.hub.Join(room, client)

	go client.WritePump()
	go client.ReadPump()
}
