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

// brokers artık []string alıyor
func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	return &KafkaProducer{writer: writer}
}

// ChatEvent → artık conversation_id ve timestamp içeriyor
type ChatEvent struct {
	EventType      string `json:"event_type"`      // "chat_completed"
	UserID         string `json:"user_id"`         // kullanıcı id'si
	Message        string `json:"message"`         // kullanıcının gönderdiği mesaj
	Response       string `json:"response"`        // AI cevabı
	ConversationID string `json:"conversation_id"` // conversation id
	Timestamp      string `json:"timestamp"`       // ISO8601 / RFC3339 formatlı zaman
	DecreaseQuota  bool   `json:"decrease_quota"`  // event-driven kota azaltma için
}

// PublishChatCompleted: Chat tamamlandığında Kafka event üretir
// conversationID "" ise event içinde boş gönderilir (consumer 'default' atayabilir)
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
