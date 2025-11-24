package handler

import (
	"net/http"

	"chat_service/internal/config"
	"chat_service/internal/models"
	"chat_service/internal/services"

	"github.com/gofiber/fiber/v2"
)

type ChatHandler struct {
	chatService *services.ChatService
}

func NewChatHandler(cfg *config.Config) *ChatHandler {
	return &ChatHandler{
		chatService: services.NewChatService(cfg),
	}
}

func (h *ChatHandler) HandleChat(c *fiber.Ctx) error {
	var req models.ChatRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.UserID == "" || req.Message == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "user_id and message are required",
		})
	}

	response, convID, err := h.chatService.HandleUserMessage(
		req.UserID,
		req.Message,
		req.ConversationID,
		req.FileID, // YENİ
	)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"response":        response,
		"conversation_id": convID,
	})
}

func (h *ChatHandler) GetFileStatus(c *fiber.Ctx) error {
	fileID := c.Params("file_id")
	if fileID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "file_id required",
		})
	}

	status := h.chatService.GetFileTracker().GetStatus(fileID)
	if status == nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "file not found",
		})
	}

	return c.JSON(status)
}
