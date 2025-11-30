package router

import (
	"net/http"

	"api_gateway/internal/handler"
)

func RegisterChatRoutes(mux *http.ServeMux, chatProxy *handler.ProxyHandler) {
	mux.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		chatProxy.ServeHTTP(w, r)
	})
}
