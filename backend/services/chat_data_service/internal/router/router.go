package router

import (
	"net/http"

	"chat_data_service/internal/handler"
	"chat_data_service/internal/services"
)

func SetupRouter(service *services.ChatDataService) *http.ServeMux {
	r := http.NewServeMux()
	h := handler.NewChatDataHandler(service)

	r.HandleFunc("/api/chat/history/", h.GetHistory)
	return r
}
