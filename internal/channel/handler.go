package channel

import (
	"net/http"

	"github.com/AzizovHikmatullo/j-support/internal/contacts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type handler struct {
	contactService contacts.Service
}

func NewHandler(service contacts.Service) *handler {
	return &handler{
		contactService: service,
	}
}

func (h *handler) InitWeb(c *gin.Context) {
	var req InitWebRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	sessionID := "js_" + uuid.Must(uuid.NewV7()).String()

	contact, err := h.contactService.InitContact(c.Request.Context(), sessionID, req.Name, req.Phone)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create contact"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contact": contact,
	})
}

func (h *handler) InitTelegram(c *gin.Context) {
	var req InitTelegramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	contact, err := h.contactService.InitContact(c.Request.Context(), req.TelegramID, req.Name, req.Phone)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create contact"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contact": contact,
	})
}
