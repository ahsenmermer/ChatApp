package router

import (
	"log"
	"net/http"

	"api_gateway/internal/config"
	"api_gateway/internal/handler"
	"api_gateway/internal/middleware"
)

func SetupRoutes(cfg config.Config) http.Handler {
	mux := http.NewServeMux()

	// Health Check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"api_gateway"}`))
	})

	// Create proxy handlers for each service
	authProxy := handler.NewProxyHandler(cfg.AuthServiceURL, "Auth")
	subscriptionProxy := handler.NewProxyHandler(cfg.SubscriptionServiceURL, "Subscription")
	chatProxy := handler.NewProxyHandler(cfg.ChatServiceURL, "Chat")
	chatDataProxy := handler.NewProxyHandler(cfg.ChatDataServiceURL, "ChatData")
	ocrProxy := handler.NewProxyHandler(cfg.OCRServiceURL, "OCR")
	embeddingProxy := handler.NewProxyHandler(cfg.EmbeddingServiceURL, "Embedding")

	// Register routes
	RegisterAuthRoutes(mux, authProxy)
	RegisterSubscriptionRoutes(mux, subscriptionProxy)
	RegisterChatRoutes(mux, chatProxy)
	RegisterChatDataRoutes(mux, chatDataProxy)
	RegisterOCRRoutes(mux, ocrProxy)
	RegisterEmbeddingRoutes(mux, embeddingProxy)

	// Fallback - 404
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("⚠️ Unknown endpoint: %s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"endpoint not found"}`))
	})

	// Apply CORS Middleware
	return middleware.Cors(mux)
}
