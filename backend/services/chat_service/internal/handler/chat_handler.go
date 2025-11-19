package handler

import (
	"net/http"

	"chat_service/internal/config"
	"chat_service/internal/models"
	"chat_service/internal/services"

	"github.com/gofiber/fiber/v2"
)

// ChatHandler HTTP isteklerini ChatService'e yönlendirir
type ChatHandler struct {
	chatService *services.ChatService
}

// NewChatHandler: ChatHandler'ı başlatır
func NewChatHandler(cfg *config.Config) *ChatHandler {
	return &ChatHandler{
		chatService: services.NewChatService(cfg),
	}
}

func (h *ChatHandler) HandleChat(c *fiber.Ctx) error {
	var req models.ChatRequest

	// Body parse
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Basic validation
	if req.UserID == "" || req.Message == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "user_id and message are required",
		})
	}

	// 🔥 Artık servis conversationID oluşturuyor ve geri döndürüyor
	response, convID, err := h.chatService.HandleUserMessage(
		req.UserID,
		req.Message,
		req.ConversationID, // frontend boş gönderebilir
	)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Frontend'e AI cevabıyla birlikte conversation_id döndür
	return c.JSON(fiber.Map{
		"response":        response,
		"conversation_id": convID,
	})
}
