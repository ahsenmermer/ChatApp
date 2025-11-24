package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
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
		client:  &http.Client{Timeout: 10 * time.Second},
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
		return nil, fmt.Errorf("qdrant search responded %d", resp.StatusCode)
	}

	var result struct {
		Result []SearchResult `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Result, nil
}
