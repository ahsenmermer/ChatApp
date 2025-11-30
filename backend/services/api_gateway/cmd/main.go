package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"api_gateway/internal/config"
	"api_gateway/internal/middleware"
	"api_gateway/internal/router"
)

func main() {
	// load env-based config
	cfg := config.Load()

	// optional: simple logging to file if requested
	if f := os.Getenv("GATEWAY_LOG_FILE"); f != "" {
		logFile, err := os.OpenFile(f, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			log.SetOutput(logFile)
			defer logFile.Close()
		} else {
			log.Printf("⚠️ Failed to open log file %s: %v", f, err)
		}
	}

	// Setup routes (CORS içeride uygulanıyor)
	mux := router.SetupRoutes(cfg)

	// ✅ Sadece logger ekle (CORS zaten router'da)
	handler := middleware.RequestLogger(mux)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("🚪 API Gateway starting on :%s (proxying: auth=%s, sub=%s, chat=%s, chatdata=%s)",
		cfg.Port, cfg.AuthServiceURL, cfg.SubscriptionServiceURL, cfg.ChatServiceURL, cfg.ChatDataServiceURL)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ API Gateway failed: %v", err)
	}
}
