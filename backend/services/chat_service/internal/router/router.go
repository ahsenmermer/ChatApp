package router

import (
	"chat_service/internal/handler"
	"chat_service/internal/services"

	"github.com/gofiber/fiber/v2"
)

// ✅ ChatService'i parametre olarak al
func SetupRouter(chatSvc *services.ChatService) *fiber.App {
	app := fiber.New()

	// ✅ Handler'a mevcut ChatService'i geçir
	chatHandler := handler.NewChatHandlerWithService(chatSvc)

	api := app.Group("/api")
	api.Post("/chat", chatHandler.HandleChat)
	api.Get("/file/status/:file_id", chatHandler.GetFileStatus)

	return app
}
