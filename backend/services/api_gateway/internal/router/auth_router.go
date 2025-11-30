package router

import (
	"net/http"

	"api_gateway/internal/handler"
)

func RegisterAuthRoutes(mux *http.ServeMux, authProxy *handler.ProxyHandler) {
	// Register endpoint - path değiştirmeye gerek yok
	mux.HandleFunc("/api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		authProxy.ServeHTTP(w, r)
	})

	// Login endpoint
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		authProxy.ServeHTTP(w, r)
	})
}
