package router

import (
	"chat_service/internal/config"
	"chat_service/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func SetupRouter(cfg *config.Config) *fiber.App {
	app := fiber.New()

	chatHandler := handler.NewChatHandler(cfg)

	api := app.Group("/api")
	api.Post("/chat", chatHandler.HandleChat)
	api.Get("/file/status/:file_id", chatHandler.GetFileStatus) // YENİ

	return app
}
