package router

import (
	"net/http"

	"api_gateway/internal/config"
	"api_gateway/internal/middleware"
	"api_gateway/internal/proxy"
)

// SetupRoutes builds the mux and wires proxies to paths.
func SetupRoutes(cfg config.Config) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"api_gateway"}`))
	})

	// Proxies
	mux.Handle("/api/auth/", proxy.NewAuthProxy(cfg.AuthServiceURL))
	mux.Handle("/api/subscription/", proxy.NewSubscriptionProxy(cfg.SubscriptionServiceURL))
	mux.Handle("/api/chat/", proxy.NewChatProxy(cfg.ChatServiceURL))
	mux.Handle("/api/chat/history/", proxy.NewChatDataProxy(cfg.ChatDataServiceURL))

	// Upload endpoint
	mux.Handle("/api/upload", UploadHandler(cfg)) // artık aynı package içinde

	// Fallback
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	return middleware.Cors(mux)
}
