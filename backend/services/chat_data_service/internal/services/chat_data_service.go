package services

import (
	"context"
	"log"
	"time"

	"chat_data_service/internal/models"
	"chat_data_service/internal/repository"
)

// parseTimestamp Kafka'dan gelen RFC3339 formatındaki timestamp'ı Go time.Time'a çevirir
func parseTimestamp(ts string) time.Time {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		log.Printf("⚠️ Timestamp parse hatası: %v", err)
		return time.Now() // varsayılan olarak şimdiki zamanı atar
	}
	return t
}

type ChatDataService struct {
	repo *repository.ChatRepository
}

func NewChatDataService(r *repository.ChatRepository) *ChatDataService {
	return &ChatDataService{repo: r}
}

// SaveMessage Kafka'dan gelen timestamp'ı ve conversation_id'yi kullanır
func (s *ChatDataService) SaveMessage(msg *models.ChatMessage) error {
	ctx := context.Background()

	// Eğer conversation_id boşsa default değer atayabiliriz
	if msg.ConversationID == "" {
		msg.ConversationID = "default"
	}

	return s.repo.SaveMessage(ctx, msg)
}

// GetHistory kullanıcının mesaj geçmişini getirir
func (s *ChatDataService) GetHistory(userID string, conversationID string, limit int, since time.Time) ([]models.ChatMessage, error) {
	ctx := context.Background()
	return s.repo.GetHistory(ctx, userID, conversationID, limit, since)
}
