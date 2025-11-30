package models

import "time"

// Frontend'den gelen istek
type ChatRequest struct {
	UserID         string `json:"user_id"`
	Message        string `json:"message"`
	ConversationID string `json:"conversation_id,omitempty"`
	FileID         string `json:"file_id,omitempty"` // ✅ YENİ - RAG için
}

// Veritabanı / Kafka / servislerde kullanılacak mesaj modeli
type ChatMessage struct {
	UserID         string    `json:"user_id"`
	Message        string    `json:"message"`
	Response       string    `json:"response"`
	ConversationID string    `json:"conversation_id,omitempty"`
	FileID         string    `json:"file_id,omitempty"` // ✅ YENİ - RAG için
	Timestamp      time.Time `json:"timestamp,omitempty"`
}
