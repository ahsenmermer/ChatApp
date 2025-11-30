package router

import (
	"net/http"

	"auth_service/internal/handler"
)

// SetupAuthRoutes sets up all routes for the auth service
func SetupAuthRoutes(mux *http.ServeMux, authHandler *handler.AuthHandler) {
	// User registration
	mux.HandleFunc("/api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandler.Register(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandler.Login(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
}
