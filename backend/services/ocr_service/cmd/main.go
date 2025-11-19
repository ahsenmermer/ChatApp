package main

import (
	"ocr_service/internal/config"
	"ocr_service/internal/router"
)

func main() {
	cfg := config.LoadConfig()
	r := router.SetupRouter()
	r.Run(":" + cfg.Port)
}
