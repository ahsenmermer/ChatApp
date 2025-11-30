package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type QdrantClient struct {
	baseURL string
	client  *http.Client
}

func NewQdrantClient(baseURL string) *QdrantClient {
	return &QdrantClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type SearchResult struct {
	ID      string                 `json:"id"`
	Score   float64                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
}

func (q *QdrantClient) Search(collection string, vector []float32, limit int, filter map[string]interface{}) ([]SearchResult, error) {
	url := fmt.Sprintf("%s/collections/%s/points/search", q.baseURL, collection)

	body := map[string]interface{}{
		"vector":       vector,
		"limit":        limit,
		"with_payload": true,
		"with_vector":  false,
	}

	if filter != nil {
		body["filter"] = filter
	}

	b, _ := json.Marshal(body)
	resp, err := q.client.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("qdrant search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Qdrant error: Status=%d, Response=%s", resp.StatusCode, string(bodyBytes))
		return nil, fmt.Errorf("qdrant search responded %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Result []SearchResult `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("✅ Qdrant returned %d results", len(result.Result))
	return result.Result, nil
}

func (q *QdrantClient) HealthCheck() error {
	url := fmt.Sprintf("%s/health", q.baseURL)
	resp, err := q.client.Get(url)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}
	return nil
}
