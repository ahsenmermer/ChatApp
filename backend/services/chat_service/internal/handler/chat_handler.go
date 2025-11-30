package handler

import (
	"log"
	"net/http"

	"chat_service/internal/models"
	"chat_service/internal/services"

	"github.com/gofiber/fiber/v2"
)

type ChatHandler struct {
	chatService *services.ChatService
}

// ❌ Eski fonksiyonu SİL veya deprecate et
// func NewChatHandler(cfg *config.Config) *ChatHandler

// ✅ YENİ: Mevcut ChatService'i kullan
func NewChatHandlerWithService(chatService *services.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

func (h *ChatHandler) HandleChat(c *fiber.Ctx) error {
	var req models.ChatRequest

	if err := c.BodyParser(&req); err != nil {
		log.Printf("❌ Invalid request body: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.UserID == "" || req.Message == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "user_id and message are required",
		})
	}

	log.Printf("📥 Chat request: user=%s, message='%s', file_id='%s'",
		req.UserID, truncate(req.Message, 50), req.FileID)

	response, convID, err := h.chatService.HandleUserMessage(
		req.UserID,
		req.Message,
		req.ConversationID,
		req.FileID,
	)

	if err != nil {
		log.Printf("❌ Chat error: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	log.Printf("✅ Chat response sent: conversation=%s, response_length=%d",
		convID, len(response))

	return c.JSON(fiber.Map{
		"response":        response,
		"conversation_id": convID,
		"file_id":         req.FileID,
	})
}

func (h *ChatHandler) GetFileStatus(c *fiber.Ctx) error {
	fileID := c.Params("file_id")
	if fileID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "file_id required",
		})
	}

	log.Printf("📊 File status request: file_id=%s", fileID)

	status := h.chatService.GetFileTracker().GetStatus(fileID)
	if status == nil {
		log.Printf("⚠️ File not found: %s", fileID)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "file not found or not yet processed",
		})
	}

	return c.JSON(status)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
