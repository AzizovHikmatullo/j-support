package tickets

import (
	"context"
	"net/http"

	"github.com/AzizovHikmatullo/j-support/internal/channel"
	"github.com/AzizovHikmatullo/j-support/internal/contacts"
	"github.com/AzizovHikmatullo/j-support/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, contactID int, role string, source string, req CreateTicketRequest) (*Ticket, error)
	Get(ctx context.Context, role string, userID int) ([]Ticket, error)
	GetByID(ctx context.Context, userID int, role string, ticketID uuid.UUID) (Ticket, error)
	GetMine(ctx context.Context, contactID int, ticketID uuid.UUID) (Ticket, error)
	ChangeAssigned(ctx context.Context, userID int, role string, ticketID uuid.UUID, assignedTo int) (Ticket, error)
	ChangeStatus(ctx context.Context, userID int, role string, ticketID uuid.UUID, status string) error
	RateTicket(ctx context.Context, contactID int, ticketID uuid.UUID, req CreateRatingRequest) (Rating, error)
	CreateMessage(ctx context.Context, ticketID uuid.UUID, senderID int, senderType, content string) (*Message, error)
	GetMessages(ctx context.Context, userID int, role string, ticketID uuid.UUID) ([]Message, error)
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	messages, err := h.service.GetMessages(c.Request.Context(), contact.ID, userRole, ticketID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	messages, err := h.service.GetMessages(c.Request.Context(), userID, role, ticketID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}
