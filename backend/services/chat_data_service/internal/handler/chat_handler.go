package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"chat_data_service/internal/services"

	"github.com/google/uuid"
)

type ChatDataHandler struct {
	service *services.ChatDataService
}

func NewChatDataHandler(srv *services.ChatDataService) *ChatDataHandler {
	return &ChatDataHandler{service: srv}
}

func (h *ChatDataHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	parts := splitPath(r.URL.Path)
	if len(parts) < 5 {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}
	userID := parts[4]
	if _, err := uuid.Parse(userID); err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			limit = i
		}
	}

	var since time.Time
	if v := r.URL.Query().Get("since"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			since = t
		}
	}

	// ✅ ConversationID query param ile alınır
	conversationID := r.URL.Query().Get("conversation_id")

	hist, err := h.service.GetHistory(userID, conversationID, limit, since)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, hist)
}

func splitPath(p string) []string {
	out := []string{}
	cur := ""
	for _, r := range p {
		if r == '/' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	out = append(out, cur)
	return out
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
