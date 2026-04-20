package channel

import (
	"net/http"

	"github.com/AzizovHikmatullo/j-support/internal/contacts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type handler struct {
	contactService contacts.Service
}

func NewHandler(service contacts.Service) *handler {
	return &handler{
		contactService: service,
	}
}

// @Summary      Инициализация веб-виджета (создание контакта)
// @Tags         init
// @Accept       json
// @Produce      json
// @Param        body  body  channel.InitWebRequest  true  "Имя и телефон"
// @Success      200   {object}  map[string]interface{} "contact"
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /init/web [post]
func (h *handler) InitWeb(c *gin.Context) {
	var req InitWebRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	sessionID := "js_" + uuid.Must(uuid.NewV7()).String()

	contact, err := h.contactService.InitContact(c.Request.Context(), sessionID, req.Name, req.Phone)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create contact"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contact": contact,
	})
}

// @Summary      Инициализация Telegram-бота (создание контакта)
// @Tags         init
// @Accept       json
// @Produce      json
// @Param        body  body  channel.InitTelegramRequest  true  "telegram_id, имя и телефон"
// @Success      200   {object}  map[string]interface{} "contact"
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /init/telegram [post]
func (h *handler) InitTelegram(c *gin.Context) {
	var req InitTelegramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	contact, err := h.contactService.InitContact(c.Request.Context(), req.TelegramID, req.Name, req.Phone)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create contact"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contact": contact,
	})
}
