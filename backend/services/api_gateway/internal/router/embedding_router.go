package router

import (
	"net/http"

	"api_gateway/internal/handler"
)

func RegisterEmbeddingRoutes(mux *http.ServeMux, embProxy *handler.ProxyHandler) {
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		embProxy.ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/search/file", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		embProxy.ServeHTTP(w, r)
	})
}
