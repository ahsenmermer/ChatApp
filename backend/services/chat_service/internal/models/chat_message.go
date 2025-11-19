package models

import "time"

// Frontend'den gelen istek
type ChatRequest struct {
	UserID         string `json:"user_id"`
	Message        string `json:"message"`
	ConversationID string `json:"conversation_id,omitempty"` // yeni alan
}

// Veritabanı / Kafka / servislerde kullanılacak mesaj modeli
type ChatMessage struct {
	UserID         string    `json:"user_id"`
	Message        string    `json:"message"`                   // user message
	Response       string    `json:"response"`                  // ai response
	ConversationID string    `json:"conversation_id,omitempty"` // yeni alan
	Timestamp      time.Time `json:"timestamp,omitempty"`       // opsiyonel zaman
}
