package main

import (
	"log"
	"net/http"
	"time"

	"api_gateway/internal/config"
	"api_gateway/internal/middleware"
	"api_gateway/internal/router"
)

func main() {
	cfg := config.Load()

	mux := router.SetupRoutes(cfg)
	handler := middleware.RequestLogger(mux)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("🚪 API Gateway starting on :%s", cfg.Port)
	log.Printf("   Auth:         %s", cfg.AuthServiceURL)
	log.Printf("   Subscription: %s", cfg.SubscriptionServiceURL)
	log.Printf("   Chat:         %s", cfg.ChatServiceURL)
	log.Printf("   ChatData:     %s", cfg.ChatDataServiceURL)
	log.Printf("   OCR:          %s", cfg.OCRServiceURL)
	log.Printf("   Embedding:    %s", cfg.EmbeddingServiceURL)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ API Gateway failed: %v", err)
	}
}
