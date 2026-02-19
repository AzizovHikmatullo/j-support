package categories

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Service interface {
	Create(ctx context.Context, role, name, destination string) (Category, error)
	Get(ctx context.Context, role string) ([]Category, error)
	Update(ctx context.Context, role string, id int, name string, enabled bool) (Category, error)
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
	var req CreateCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	role := c.GetString("role")

	category, err := h.service.Create(c.Request.Context(), role, req.Name, req.Destination)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

func (h *handler) Get(c *gin.Context) {
	role := c.GetString("role")

	categories, err := h.service.Get(c.Request.Context(), role)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

func (h *handler) Update(c *gin.Context) {
	var req UpdateCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	role := c.GetString("role")

	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
	}

	category, err := h.service.Update(c.Request.Context(), role, idInt, req.Name, req.Enabled)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}
