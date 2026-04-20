package tickets

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/AzizovHikmatullo/j-support/internal/channel"
	"github.com/AzizovHikmatullo/j-support/internal/contacts"
	"github.com/AzizovHikmatullo/j-support/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, contactID int, role string, source string, req CreateTicketRequest) (*CreateTicketResponse, error)
	Get(ctx context.Context, role string, userID int) ([]Ticket, error)
	GetByID(ctx context.Context, userID int, role string, ticketID uuid.UUID) (Ticket, error)
	GetMine(ctx context.Context, contactID int, ticketID uuid.UUID) (Ticket, error)
	ChangeAssigned(ctx context.Context, userID int, role string, ticketID uuid.UUID, assignedTo int) (Ticket, error)
	ChangeStatus(ctx context.Context, userID int, role string, ticketID uuid.UUID, status string) error
	RateTicket(ctx context.Context, contactID int, ticketID uuid.UUID, req CreateRatingRequest) (Rating, error)
	CreateMessage(ctx context.Context, ticketID uuid.UUID, senderID int, senderType, content string) (*Message, error)
	CreateMessageWithButtons(ctx context.Context, ticketID uuid.UUID, senderID int, senderType, content string, buttons []string) (*Message, error)
	GetMessages(ctx context.Context, userID int, role string, ticketID uuid.UUID, limit int, cursor string) ([]Message, string, error)
	SetScenarioService(botService scenarioService)
}

type handler struct {
	service  Service
	registry *channel.Registry
}

func NewHandler(service Service, registry *channel.Registry) *handler {
	return &handler{
		service:  service,
		registry: registry,
	}
}

// CLIENT HANDLERS

func (h *handler) Create(c *gin.Context) {
	var req CreateTicketRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	identity, ok := middleware.GetIdentity(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorized.Error()})
		return
	}

	contact, err := h.resolveContact(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ticket, err := h.service.Create(c.Request.Context(), contact.ID, userRole, identity.ChannelType, req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

func (h *handler) GetMine(c *gin.Context) {
	contact, err := h.resolveContact(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	tickets, err := h.service.Get(c.Request.Context(), userRole, contact.ID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, tickets)
}

func (h *handler) GetMineByID(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid ticketID"})
		return
	}

	contact, err := h.resolveContact(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ticket, err := h.service.GetMine(c.Request.Context(), contact.ID, ticketID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func (h *handler) Rate(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid ticketID"})
		return
	}

	var req CreateRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	contact, err := h.resolveContact(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	rating, err := h.service.RateTicket(c.Request.Context(), contact.ID, ticketID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, rating)
}

func (h *handler) CreateMessageByUser(c *gin.Context) {
	var req CreateMessageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid ticketID"})
		return
	}

	contact, err := h.resolveContact(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	message, err := h.service.CreateMessage(c.Request.Context(), ticketID, contact.ID, userRole, req.Content)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, message)
}

func (h *handler) GetMessagesForUser(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid ticketID"})
		return
	}

	contact, err := h.resolveContact(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	cursor := c.Query("cursor")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	messages, nextCursor, err := h.service.GetMessages(c.Request.Context(), contact.ID, userRole, ticketID, limit, cursor)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages, "nextCursor": nextCursor})
}

func (h *handler) resolveContact(c *gin.Context) (*contacts.Contact, error) {
	identity, ok := middleware.GetIdentity(c)
	if !ok {
		return nil, ErrUnauthorized
	}

	ch, err := h.registry.Get(identity.ChannelType)
	if err != nil {
		return nil, ErrUnknownChannel
	}

	return ch.ResolveContact(c.Request.Context(), identity.ID)
}

// SUPPORT HANDLERS

func (h *handler) Get(c *gin.Context) {
	role := c.GetString("role")
	userID := c.GetInt("userID")

	tickets, err := h.service.Get(c.Request.Context(), role, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, tickets)
}

func (h *handler) GetByID(c *gin.Context) {
	role := c.GetString("role")
	userID := c.GetInt("userID")

	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid ticketID"})
		return
	}

	ticket, err := h.service.GetByID(c.Request.Context(), userID, role, ticketID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func (h *handler) ChangeAssigned(c *gin.Context) {
	var req ChangeAssignedRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid ticketID"})
		return
	}

	role := c.GetString("role")
	userID := c.GetInt("userID")

	ticket, err := h.service.ChangeAssigned(c.Request.Context(), userID, role, ticketID, req.AssignedTo)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func (h *handler) ChangeStatus(c *gin.Context) {
	var req ChangeStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid ticketID"})
		return
	}

	role := c.GetString("role")
	userID := c.GetInt("userID")

	err = h.service.ChangeStatus(c.Request.Context(), userID, role, ticketID, req.Status)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *handler) CreateMessageBySupport(c *gin.Context) {
	var req CreateMessageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid ticketID"})
		return
	}

	role := c.GetString("role")
	userID := c.GetInt("userID")

	message, err := h.service.CreateMessage(c.Request.Context(), ticketID, userID, role, req.Content)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, message)
}

func (h *handler) GetMessagesForSupport(c *gin.Context) {
	role := c.GetString("role")
	userID := c.GetInt("userID")

	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid ticketID"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	cursor := c.Query("cursor")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	messages, nextCursor, err := h.service.GetMessages(c.Request.Context(), userID, role, ticketID, limit, cursor)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages, "nextCursor": nextCursor})
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrForbidden):
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": ErrForbidden.Error()})
	case errors.Is(err, ErrUnauthorized):
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorized.Error()})
	case errors.Is(err, ErrRatingNotFound):
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": ErrRatingNotFound.Error()})
	case errors.Is(err, ErrTicketNotFound):
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": ErrTicketNotFound.Error()})
	case errors.Is(err, ErrUnknownChannel):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": ErrUnknownChannel.Error()})
	case errors.Is(err, ErrInvalidStatus):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": ErrInvalidStatus.Error()})
	case errors.Is(err, ErrInvalidScore):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": ErrInvalidScore.Error()})
	case errors.Is(err, ErrClosedTicket):
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": ErrClosedTicket.Error()})
	case errors.Is(err, ErrNotClosed):
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": ErrNotClosed.Error()})
	case errors.Is(err, ErrCategoryDisabled):
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": ErrCategoryDisabled.Error()})
	case errors.Is(err, ErrCannotAssign):
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": ErrCannotAssign.Error()})
	case errors.Is(err, ErrSupportCannotWrite):
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": ErrSupportCannotWrite.Error()})
	case errors.Is(err, ErrAlreadyRated):
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": ErrRatingNotFound.Error()})
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
