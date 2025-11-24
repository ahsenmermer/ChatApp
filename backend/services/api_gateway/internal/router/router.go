package router

import (
	"net/http"

	"api_gateway/internal/config"
	"api_gateway/internal/handler"
	"api_gateway/internal/middleware"
)

func SetupRoutes(cfg config.Config) http.Handler {
	mux := http.NewServeMux()

	// Auth Service Routes
	mux.Handle("/api/auth/", handler.NewAuthHandler(cfg.AuthServiceURL))

	// Subscription Service Routes
	mux.Handle("/api/subscription/", handler.NewSubscriptionHandler(cfg.SubscriptionServiceURL))

	// Chat Service Routes
	mux.Handle("/api/chat", handler.NewChatHandler(cfg.ChatServiceURL))
	mux.Handle("/api/file/status/", handler.NewChatHandler(cfg.ChatServiceURL))

	// Chat Data Service Routes
	mux.Handle("/api/chat/history/", handler.NewChatDataHandler(cfg.ChatDataServiceURL))

	// OCR Service Routes (dosya yükleme)
	mux.HandleFunc("/api/upload", func(w http.ResponseWriter, r *http.Request) {
		// OCR servisine yönlendir
		proxy := handler.NewOCRHandler(cfg.OCRServiceURL)
		proxy.ServeHTTP(w, r)
	})

	// Fallback
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	return middleware.Cors(mux)
}
