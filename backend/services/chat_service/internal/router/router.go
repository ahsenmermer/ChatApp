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
	// Kullanıcı her zaman chat yapabilir, plan kontrolü yok
	api.Post("/chat", chatHandler.HandleChat)

	return app
}
