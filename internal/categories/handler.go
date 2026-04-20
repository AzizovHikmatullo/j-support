package categories

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/AzizovHikmatullo/j-support/internal/middleware"
	"github.com/gin-gonic/gin"
)

type Service interface {
	Create(ctx context.Context, name, destination string) (Category, error)
	Get(ctx context.Context, role string) ([]Category, error)
	Update(ctx context.Context, id int, name *string, enabled *bool) (Category, error)
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

// @Summary      Создать новую категорию
// @Description  Только администратор
// @Tags         categories
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        body  body  categories.CreateCategoryRequest  true  "Создание категории"
// @Success      201   {object}  categories.Category
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /categories [post]
func (h *handler) Create(c *gin.Context) {
	var req CreateCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	category, err := h.service.Create(c.Request.Context(), req.Name, req.Destination)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, category)
}

// @Summary      Получить список категорий
// @Description  Для user/web/telegram — только активные категории по destination. Для admin/support — все категории.
// @Tags         categories
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {array}   categories.Category
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /categories [get]
func (h *handler) Get(c *gin.Context) {
	identity, ok := middleware.GetIdentity(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorized.Error()})
		return
	}

	categories, err := h.service.Get(c.Request.Context(), identity.Role)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, categories)
}

// @Summary      Обновить категорию
// @Description  Только администратор
// @Tags         categories
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id    path      int                              true  "ID категории"
// @Param        body  body      categories.UpdateCategoryRequest true  "Обновление"
// @Success      200    {object}  categories.Category
// @Failure      400    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /categories/{id} [patch]
func (h *handler) Update(c *gin.Context) {
	var req UpdateCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	category, err := h.service.Update(c.Request.Context(), idInt, req.Name, req.Enabled)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, category)
}

func (h *handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidDest), errors.Is(err, ErrInvalidName):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, ErrUnauthorized):
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		h.logger.Error("category error", "error", err.Error())
	}
}
