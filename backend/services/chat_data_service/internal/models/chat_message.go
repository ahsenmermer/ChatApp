package models

import "time"

type ChatMessage struct {
	UserID         string    `ch:"user_id" json:"user_id"`
	UserMessage    string    `ch:"user_message" json:"user_message"`
	AIResponse     string    `ch:"ai_response" json:"ai_response"`
	ConversationID string    `ch:"conversation_id" json:"conversation_id"`
	Timestamp      time.Time `ch:"timestamp" json:"timestamp"`
}
