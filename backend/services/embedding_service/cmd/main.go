package main

import (
	"embedding_service/internal/config"
	"embedding_service/internal/router"
)

func main() {
	cfg := config.LoadConfig()
	r := router.SetupRouter()
	r.Run(":" + cfg.Port)
}
