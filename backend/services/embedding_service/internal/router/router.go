package router

import (
	"log"
	"net/http"

	"embedding_service/internal/services"
)

// SetupRouter configures all HTTP routes
func SetupRouter(searchSvc *services.SearchService) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"embedding"}`))
	})

	// Search endpoints
	mux.HandleFunc("/api/search", searchSvc.HandleSearch)
	mux.HandleFunc("/api/search/file", searchSvc.HandleSearchByFileID)

	// CORS middleware
	corsHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			h.ServeHTTP(w, r)
		})
	}

	log.Printf("🌐 Router configured successfully")
	return corsHandler(mux)
}
