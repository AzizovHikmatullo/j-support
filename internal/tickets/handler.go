package tickets

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, userID int, role string, req CreateTicketRequest) (*Ticket, error)
	Get(ctx context.Context, role string, userID int) ([]Ticket, error)
	GetByID(ctx context.Context, userID int, role string, ticketID uuid.UUID) (Ticket, error)
	ChangeAssigned(ctx context.Context, userID int, role string, ticketID uuid.UUID, assignedTo int) (Ticket, error)
	ChangeStatus(ctx context.Context, userID int, role string, ticketID uuid.UUID, status string) (Ticket, error)
	CreateMessage(ctx context.Context, ticketID uuid.UUID, senderID int, senderType, content string) (*Message, error)
	GetMessages(ctx context.Context, userID int, role string, ticketID uuid.UUID) ([]Message, error)
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
	var req CreateTicketRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	ticket, err := h.service.Create(c.Request.Context(), c.GetInt("userID"), c.GetString("role"), req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

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

	ticket, err := h.service.ChangeStatus(c.Request.Context(), userID, role, ticketID, req.Status)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func (h *handler) CreateMessage(c *gin.Context) {
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

func (h *handler) GetMessages(c *gin.Context) {
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
