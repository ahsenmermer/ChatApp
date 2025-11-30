package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	return &KafkaProducer{writer: writer}
}

type ChatEvent struct {
	EventType      string `json:"event_type"`
	UserID         string `json:"user_id"`
	Message        string `json:"message"`
	Response       string `json:"response"`
	ConversationID string `json:"conversation_id"`
	Timestamp      string `json:"timestamp"`
	DecreaseQuota  bool   `json:"decrease_quota"`
}

func (k *KafkaProducer) PublishChatCompleted(userID, message, response, conversationID string) error {
	event := ChatEvent{
		EventType:      "chat_completed",
		UserID:         userID,
		Message:        message,
		Response:       response,
		ConversationID: conversationID,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		DecreaseQuota:  true,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = k.writer.WriteMessages(context.Background(),
		kafka.Message{
			Value: data,
		})
	if err != nil {
		fmt.Println("Kafka publish error:", err)
	}
	return err
}

// ✅ YENİ: PublishFileAttached - PDF yükleme mesajını kaydet
type FileAttachedEvent struct {
	EventType      string `json:"event_type"` // "file_attached"
	UserID         string `json:"user_id"`
	Message        string `json:"message"` // "📎 öneri.pdf"
	ConversationID string `json:"conversation_id"`
	FileID         string `json:"file_id"`
	FileName       string `json:"file_name"`
	Timestamp      string `json:"timestamp"`
}

func (k *KafkaProducer) PublishFileAttached(userID, fileName, fileID, conversationID string) error {
	event := FileAttachedEvent{
		EventType:      "file_attached",
		UserID:         userID,
		Message:        fmt.Sprintf("📎 %s", fileName),
		ConversationID: conversationID,
		FileID:         fileID,
		FileName:       fileName,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = k.writer.WriteMessages(context.Background(),
		kafka.Message{
			Value: data,
		})
	if err != nil {
		fmt.Println("Kafka publish error:", err)
	}
	return err
}
