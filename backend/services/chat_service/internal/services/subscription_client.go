package services

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/IBM/sarama"
)

// SubscriptionClient: Subscription Service ile iletişimi sağlar
type SubscriptionClient struct {
	BaseURL      string
	KafkaBrokers []string
	KafkaTopic   string
}

type QuotaResponse struct {
	Quota int `json:"quota"`
}

func NewSubscriptionClient(baseURL string, brokers []string, topic string) *SubscriptionClient {
	return &SubscriptionClient{
		BaseURL:      baseURL,
		KafkaBrokers: brokers,
		KafkaTopic:   topic,
	}
}

// Kullanıcının aktif planı var mı?
func (s *SubscriptionClient) IsSubscriptionActive(userID string) bool {
	resp, err := http.Get(fmt.Sprintf("%s/api/subscription/quota/%s", s.BaseURL, userID))
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	defer resp.Body.Close()

	var result QuotaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}

	return result.Quota > 0
}

// Kullanıcının kalan kotasını döner
func (s *SubscriptionClient) GetQuota(userID string) (int, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/subscription/quota/%s", s.BaseURL, userID))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get quota, status: %d", resp.StatusCode)
	}

	var result QuotaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result.Quota, nil
}

// (İsteğe bağlı) Kafka üzerinden kota azaltma event'i gönderir
func (s *SubscriptionClient) DecreaseQuota(userID string) error {
	producer, err := sarama.NewSyncProducer(s.KafkaBrokers, nil)
	if err != nil {
		return fmt.Errorf("failed to create Kafka producer: %w", err)
	}
	defer producer.Close()

	event := map[string]string{
		"type":    "chat_completed",
		"user_id": userID,
	}

	data, _ := json.Marshal(event)
	msg := &sarama.ProducerMessage{
		Topic: s.KafkaTopic,
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send Kafka message: %w", err)
	}

	return nil
}
