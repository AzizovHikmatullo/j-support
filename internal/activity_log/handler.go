package activity_log

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Service interface {
	Log(ctx context.Context, entry LogEntry)
	GetByTicket(ctx context.Context, ticketID uuid.UUID) ([]ActivityLog, error)
	GetAll(ctx context.Context) ([]ActivityLog, error)
}

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) GetAll(c *gin.Context) {
	logs, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, logs)
}

func (h *handler) GetByTicket(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid ticketID"})
		return
	}

	logs, err := h.service.GetByTicket(c.Request.Context(), ticketID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, logs)
}

func handleError(c *gin.Context, err error) {
	switch err {
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
