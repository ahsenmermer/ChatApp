package models

import "time"

type ChatRequest struct {
	UserID         string `json:"user_id"`
	Message        string `json:"message"`
	ConversationID string `json:"conversation_id,omitempty"`
	FileID         string `json:"file_id,omitempty"` // YENİ
}

type ChatMessage struct {
	UserID         string    `json:"user_id"`
	Message        string    `json:"message"`
	Response       string    `json:"response"`
	ConversationID string    `json:"conversation_id,omitempty"`
	FileID         string    `json:"file_id,omitempty"` // YENİ
	Timestamp      time.Time `json:"timestamp,omitempty"`
}
