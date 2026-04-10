package ws

import (
	"net/http"

	"github.com/AzizovHikmatullo/j-support/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WSHandler struct {
	hub *Hub
	cfg *config.Config
}

func NewWSHandler(hub *Hub, cfg *config.Config) *WSHandler {
	return &WSHandler{
		hub: hub,
		cfg: cfg,
	}
}

func (h *WSHandler) getUpgrader() websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")

			for _, o := range h.cfg.WS.AllowedOrigins {
				if o == origin {
					return true
				}
			}
			return false
		},
	}
}

func (h *WSHandler) ServeTicketWS(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}

	upgrader := h.getUpgrader()

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
