package scenario

import (
	"context"
	"net/http"
	"strconv"

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
	UpdateStep(ctx context.Context, scenarioID, stepID int, req UpdateStepRequest) (Step, error)
	DeleteStep(ctx context.Context, scenarioID, stepID int) error

	StartIfExists(ctx context.Context, ticketID uuid.UUID, categoryID int) error
	HandleMessage(ctx context.Context, ticketID uuid.UUID, answer string) (*string, error)
}

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, scenario)
}

func (h *handler) GetAll(c *gin.Context) {
	scenarios, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
