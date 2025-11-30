package router

import (
	"net/http"

	"api_gateway/internal/handler"
)

func RegisterOCRRoutes(mux *http.ServeMux, ocrProxy *handler.ProxyHandler) {
	mux.HandleFunc("/api/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		ocrProxy.ServeHTTP(w, r)
	})

	// File status endpoint
	mux.HandleFunc("/api/file/status/", ocrProxy.ServeHTTP)
}
