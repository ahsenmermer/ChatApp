package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"chat_service/internal/config"
	"chat_service/internal/repository"

	"github.com/google/uuid"
)

type ChatService struct {
	cfg                *config.Config
	authClient         *AuthClient
	subscriptionClient *SubscriptionClient
	memoryService      *MemoryService
	kafkaProducer      *repository.KafkaProducer
	ragService         *RAGService
	fileTracker        *FileTracker
}

func NewChatService(cfg *config.Config) *ChatService {
	qdrantClient := repository.NewQdrantClient(cfg.QdrantURL)
	xenovaClient := NewXenovaClient(cfg.XenovaURL)

	return &ChatService{
		cfg:                cfg,
		authClient:         NewAuthClient(cfg.AuthServiceURL),
		subscriptionClient: NewSubscriptionClient(cfg.SubscriptionServiceURL, cfg.KafkaBrokers, cfg.KafkaTopicChatMessages),
		memoryService:      NewMemoryService(20),
		kafkaProducer:      repository.NewKafkaProducer(cfg.KafkaBrokers, cfg.KafkaTopicChatMessages),
		ragService:         NewRAGService(qdrantClient, xenovaClient),
		fileTracker:        NewFileTracker(),
	}
}

func (c *ChatService) HandleUserMessage(userID, message, conversationID, fileID string) (string, string, error) {
	if conversationID == "" {
		conversationID = uuid.New().String()
	}

	if !c.authClient.IsUserValid(userID) {
		return "", conversationID, fmt.Errorf("unauthorized user")
	}

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

	// RAG: Eğer fileID varsa ve dosya hazırsa, ilgili chunk'ları getir
	var ragContext string
	if fileID != "" {
		if !c.fileTracker.IsReady(fileID) {
			return "", conversationID, fmt.Errorf("file is still processing or not found")
		}

		chunks, err := c.ragService.SearchRelevantChunks(context.Background(), message, fileID, 5)
		if err != nil {
			return "", conversationID, fmt.Errorf("failed to search document: %v", err)
		}

		if len(chunks) > 0 {
			ragContext = c.ragService.BuildContext(chunks)
		}
	}

	// Memory'den geçmiş konuşmaları al
	memoryContext := c.memoryService.GetContext(userID)

	// Full prompt oluştur
	var fullPrompt string
	if ragContext != "" {
		fullPrompt = fmt.Sprintf("%s\n\n%s\n\nKullanıcı Sorusu: %s", ragContext, memoryContext, message)
	} else {
		fullPrompt = fmt.Sprintf("%s\nUser: %s", memoryContext, message)
	}

	// OpenRouter API çağrısı
	systemPrompt := "Sen kullanıcıyla doğal bir şekilde sohbet eden bir yapay zekâsın."
	if ragContext != "" {
		systemPrompt = "Sen kullanıcıya belge içeriğine dayalı cevaplar veren bir yapay zekâsın. Verilen belge bölümlerini analiz edip kullanıcının sorusuna doğru ve detaylı cevap ver."
	}

	response, err := callOpenRouterAPI(fullPrompt, systemPrompt, c.cfg.OpenRouterKey)
	if err != nil {
		return "", conversationID, fmt.Errorf("AI response error: %v", err)
	}

	// Memory'e ekle
	c.memoryService.AddMessage(userID, "User: "+message)
	c.memoryService.AddMessage(userID, "AI: "+response)

	// Subscription event gönder
	go func() {
		event := map[string]string{
			"user_id": userID,
			"event":   "message_sent",
		}
		jsonData, _ := json.Marshal(event)
		http.Post(fmt.Sprintf("%s/api/subscription/event", c.cfg.SubscriptionServiceURL),
			"application/json", bytes.NewBuffer(jsonData))
	}()

	// Kafka event
	go func() {
		if err := c.kafkaProducer.PublishChatCompleted(userID, message, response, conversationID); err != nil {
			fmt.Printf("Kafka event publish failed: %v\n", err)
		}
	}()

	return response, conversationID, nil
}

func (c *ChatService) GetFileTracker() *FileTracker {
	return c.fileTracker
}

func callOpenRouterAPI(prompt, systemPrompt, apiKey string) (string, error) {
	url := "https://openrouter.ai/api/v1/chat/completions"

	reqBody := map[string]interface{}{
		"model": "nvidia/nemotron-nano-9b-v2:free",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
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
