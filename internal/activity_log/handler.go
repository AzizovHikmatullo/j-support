package activity_log

import (
	"context"
	"log/slog"
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

	logger *slog.Logger
}

func NewHandler(service Service, logger *slog.Logger) *handler {
	return &handler{
		service: service,
		logger:  logger,
	}
}

// @Summary      Получить весь лог активности
// @Tags         activity
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {array}   activity_log.ActivityLog
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /activity [get]
func (h *handler) GetAll(c *gin.Context) {
	logs, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, logs)
}

// @Summary      Получить лог активности по тикету
// @Tags         activity
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path   string   true   "UUID тикета"
// @Success      200   {array}   activity_log.ActivityLog
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /activity/{id} [get]
func (h *handler) GetByTicket(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid ticketID"})
		return
	}

	logs, err := h.service.GetByTicket(c.Request.Context(), ticketID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, logs)
}

func (h *handler) handleError(c *gin.Context, err error) {
	switch err {
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		h.logger.Error("activity log error", "error", err.Error())
	}
}
