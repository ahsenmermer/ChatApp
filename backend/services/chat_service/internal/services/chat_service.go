package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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

	// ✅ YENİ: Eğer dosya init mesajıysa, sadece FileTracker'a kaydet ve çık
	if strings.HasPrefix(message, "_file_init_") {
		if fileID != "" {
			c.fileTracker.SetUserInfo(fileID, userID, conversationID)
			log.Printf("📝 File info registered: file=%s, user=%s, conv=%s", fileID, userID, conversationID)
		}
		return "File registered", conversationID, nil
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

	// ✅ Eğer fileID varsa, user bilgilerini FileTracker'a kaydet
	if fileID != "" {
		c.fileTracker.SetUserInfo(fileID, userID, conversationID)
	}

	// RAG - Eğer fileID varsa ve dosya hazırsa, ilgili chunk'ları getir
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

	// 3. Memory'den geçmiş konuşmaları al
	memoryContext := c.memoryService.GetContext(userID)

	// Full prompt oluştur (RAG context varsa ekle)
	var fullPrompt string
	if ragContext != "" {
		fullPrompt = fmt.Sprintf("%s\n\n%s\n\nKullanıcı Sorusu: %s", ragContext, memoryContext, message)
	} else {
		fullPrompt = fmt.Sprintf("%s\nUser: %s", memoryContext, message)
	}

	// 4. OpenRouter API çağrısı
	systemPrompt := "Sen kullanıcıyla doğal bir şekilde sohbet eden bir yapay zekâsın."
	if ragContext != "" {
		systemPrompt = "Sen kullanıcıya belge içeriğine dayalı cevaplar veren bir yapay zekâsın. Verilen belge bölümlerini analiz edip kullanıcının sorusuna doğru ve detaylı cevap ver."
	}

	response, err := callOpenRouterAPI(fullPrompt, systemPrompt, c.cfg.OpenRouterKey)
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

	// 7. Kafka'ya event gönder
	go func() {
		if err := c.kafkaProducer.PublishChatCompleted(userID, message, response, conversationID); err != nil {
			log.Printf("⚠️ Kafka event publish failed: %v", err)
		}
	}()

	return response, conversationID, nil
}

func (c *ChatService) GetFileTracker() *FileTracker {
	return c.fileTracker
}

func (c *ChatService) GetKafkaProducer() *repository.KafkaProducer {
	return c.kafkaProducer
}

func callOpenRouterAPI(prompt, systemPrompt, apiKey string) (string, error) {
	url := "https://openrouter.ai/api/v1/chat/completions"

	reqBody := map[string]interface{}{
		"model": "nvidia/nemotron-nano-9b-v2:free",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
		"max_tokens":  2000,
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://yourapp.com")
	req.Header.Set("X-Title", "ChatApp")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return "", fmt.Errorf("API returned status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode failed: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s (code: %s)", result.Error.Message, result.Error.Code)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("AI returned empty response")
	}

	return result.Choices[0].Message.Content, nil
}
