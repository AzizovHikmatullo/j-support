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
	Get(ctx context.Context, id int) (Scenario, error)
	GetAll(ctx context.Context) ([]Scenario, error)
	Update(ctx context.Context, scenarioID int, req UpdateScenarioRequest) (Scenario, error)

	StartIfExists(ctx context.Context, ticketID uuid.UUID, categoryID int) error
	HandleMessage(ctx context.Context, ticketID uuid.UUID) (*string, error)
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

func (h *handler) Get(c *gin.Context) {
	scenarioID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid scenario id"})
		return
	}

	scenario, err := h.service.Get(c.Request.Context(), scenarioID)
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
