package scenario

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/AzizovHikmatullo/j-support/internal/tickets"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Service interface {
	CreateScenario(ctx context.Context, req CreateScenarioRequest) (Scenario, error)
	GetByID(ctx context.Context, id int) (Scenario, error)
	GetAll(ctx context.Context) ([]Scenario, error)
	Update(ctx context.Context, id int, req UpdateScenarioRequest) (Scenario, error)
	Delete(ctx context.Context, id int) error

	CreateStep(ctx context.Context, scenarioID int, req CreateStepRequest) (Step, error)
	GetButtonsForCurrentStep(ctx context.Context, ticketID uuid.UUID) ([]string, error)
	UpdateStep(ctx context.Context, scenarioID, stepID int, req UpdateStepRequest) (Step, error)
	DeleteStep(ctx context.Context, scenarioID, stepID int) error

	StartIfExists(ctx context.Context, ticketID uuid.UUID, categoryID int) (*tickets.Message, []string, error)
	HandleMessage(ctx context.Context, ticketID uuid.UUID, answer string) (*string, error)
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

func (h *handler) Create(c *gin.Context) {
	var req CreateScenarioRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	scenario, err := h.service.CreateScenario(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, scenario)
}

func (h *handler) GetByID(c *gin.Context) {
	scenarioID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid scenario id"})
		return
	}

	scenario, err := h.service.GetByID(c.Request.Context(), scenarioID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, scenario)
}

func (h *handler) GetAll(c *gin.Context) {
	scenarios, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, scenarios)
}

func (h *handler) Update(c *gin.Context) {
	scenarioID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid scenario id"})
		return
	}

	var req UpdateScenarioRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	scenario, err := h.service.Update(c.Request.Context(), scenarioID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, scenario)
}

func (h *handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid scenario id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *handler) CreateStep(c *gin.Context) {
	scenarioID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid scenario id"})
		return
	}

	var req CreateStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	step, err := h.service.CreateStep(c.Request.Context(), scenarioID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, step)
}

func (h *handler) UpdateStep(c *gin.Context) {
	scenarioID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid scenario id"})
		return
	}

	stepID, err := strconv.Atoi(c.Param("stepID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid step id"})
		return
	}

	var req UpdateStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	step, err := h.service.UpdateStep(c.Request.Context(), scenarioID, stepID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, step)
}

func (h *handler) DeleteStep(c *gin.Context) {
	scenarioID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid scenario id"})
		return
	}

	stepID, err := strconv.Atoi(c.Param("stepID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid step id"})
		return
	}

	if err := h.service.DeleteStep(c.Request.Context(), scenarioID, stepID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrScenarioNotFound):
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": ErrScenarioNotFound.Error()})
	case errors.Is(err, ErrStepNotFound):
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": ErrStepNotFound.Error()})
	case errors.Is(err, ErrSessionNotFound):
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": ErrSessionNotFound.Error()})
	case errors.Is(err, ErrParentNotFound):
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": ErrParentNotFound.Error()})
	case errors.Is(err, ErrRootAlreadyExists):
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": ErrRootAlreadyExists.Error()})
	case errors.Is(err, ErrDefaultAlreadyExists):
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": ErrDefaultAlreadyExists.Error()})
	case errors.Is(err, ErrWrongScenario):
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		h.logger.Error("scenario error", "error", err.Error())
	}
}
