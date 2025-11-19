package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"chat_service/internal/config"
	"chat_service/internal/repository"

	"github.com/google/uuid"
)

// ChatService: Chat işlemlerini yöneten ana servis
type ChatService struct {
	cfg                *config.Config
	authClient         *AuthClient
	subscriptionClient *SubscriptionClient
	memoryService      *MemoryService
	kafkaProducer      *repository.KafkaProducer
}

// NewChatService: Servisi başlatır
func NewChatService(cfg *config.Config) *ChatService {
	return &ChatService{
		cfg:                cfg,
		authClient:         NewAuthClient(cfg.AuthServiceURL),
		subscriptionClient: NewSubscriptionClient(cfg.SubscriptionServiceURL, cfg.KafkaBrokers, cfg.KafkaTopicChatMessages),
		memoryService:      NewMemoryService(20),
		kafkaProducer:      repository.NewKafkaProducer(cfg.KafkaBrokers, cfg.KafkaTopicChatMessages),
	}
}

// HandleUserMessage:
// Artık 3 değer döner: response, conversationID, error
func (c *ChatService) HandleUserMessage(userID, message, conversationID string) (string, string, error) {

	// 👉 Backend conversationID oluşturuyor
	if conversationID == "" {
		conversationID = uuid.New().String()
	}

	// 1. Auth doğrulama
	if !c.authClient.IsUserValid(userID) {
		return "", conversationID, fmt.Errorf("unauthorized user")
	}

	// 2. Plan aktif mi ve kota var mı?
	if !c.subscriptionClient.IsSubscriptionActive(userID) {
		return "", conversationID, fmt.Errorf("subscription inactive or expired")
	}

	quota, err := c.subscriptionClient.GetQuota(userID)
	if err != nil {
		return "", conversationID, fmt.Errorf("quota check failed: %v", err)
	}
	if quota <= 0 {
		return "", conversationID, fmt.Errorf("quota exhausted")
	}

	// 3. Memory'den geçmiş konuşmaları al
	context := c.memoryService.GetContext(userID)
	fullPrompt := fmt.Sprintf("%s\nUser: %s", context, message)

	// 4. OpenRouter API çağrısı
	response, err := callOpenRouterAPI(fullPrompt, c.cfg.OpenRouterKey)
	if err != nil {
		return "", conversationID, fmt.Errorf("AI response error: %v", err)
	}

	// 5. Memory'e yeni mesajları ekle
	c.memoryService.AddMessage(userID, "User: "+message)
	c.memoryService.AddMessage(userID, "AI: "+response)

	// 6. Subscription Service'e kota azaltma bildirimi gönder
	go func() {
		event := map[string]string{
			"user_id": userID,
			"event":   "message_sent",
		}
		jsonData, _ := json.Marshal(event)
		http.Post(fmt.Sprintf("%s/api/subscription/event", c.cfg.SubscriptionServiceURL),
			"application/json", bytes.NewBuffer(jsonData))
	}()

	// 7. Kafka’ya event gönder
	go func() {
		if err := c.kafkaProducer.PublishChatCompleted(userID, message, response, conversationID); err != nil {
			fmt.Printf("Kafka event publish failed: %v\n", err)
		}
	}()

	// 🔥 Artık conversationID de dönüyor
	return response, conversationID, nil
}

// ============================================================
// 🧠 OpenRouter API Çağrısı
// ============================================================

func callOpenRouterAPI(prompt string, apiKey string) (string, error) {
	url := "https://openrouter.ai/api/v1/chat/completions"

	reqBody := map[string]interface{}{
		"model": "nvidia/nemotron-nano-9b-v2:free",
		"messages": []map[string]string{
			{"role": "system", "content": "Sen kullanıcıyla doğal bir şekilde sohbet eden bir yapay zekâsın."},
			{"role": "user", "content": prompt},
		},
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("AI returned empty response")
	}

	return result.Choices[0].Message.Content, nil
}
