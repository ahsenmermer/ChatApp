package repository

import (
	"context"
	"time"

	"chat_data_service/internal/models"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type ChatRepository struct {
	db clickhouse.Conn
}

func NewChatRepository(db clickhouse.Conn) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) SaveMessage(ctx context.Context, msg *models.ChatMessage) error {
	query := `INSERT INTO chat_messages (user_id, user_message, ai_response, conversation_id, timestamp) VALUES (?, ?, ?, ?, ?)`
	return r.db.Exec(ctx, query, msg.UserID, msg.UserMessage, msg.AIResponse, msg.ConversationID, msg.Timestamp)
}

func (r *ChatRepository) GetHistory(ctx context.Context, userID string, conversationID string, limit int, since time.Time) ([]models.ChatMessage, error) {
	query := `
		SELECT user_id, user_message, ai_response, conversation_id, timestamp
		FROM chat_messages
		WHERE user_id = ?
		AND timestamp >= ?
	`
	args := []interface{}{userID, since}

	if conversationID != "" && conversationID != "default" {
		query += " AND conversation_id = ?"
		args = append(args, conversationID)
	}

	query += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var msg models.ChatMessage
		if err := rows.ScanStruct(&msg); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
