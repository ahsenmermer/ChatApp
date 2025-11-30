package router

import (
	"api_gateway/internal/handler"
	"net/http"
)

func RegisterChatDataRoutes(mux *http.ServeMux, chatDataProxy *handler.ProxyHandler) {
	mux.HandleFunc("/api/chat/history/", chatDataProxy.ServeHTTP)
}
