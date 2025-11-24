package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type QdrantRepo struct {
	baseURL string
	client  *http.Client
}

func NewQdrantRepo(baseURL string) *QdrantRepo {
	return &QdrantRepo{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// UpsertVector uploads a single vector to Qdrant
func (r *QdrantRepo) UpsertVector(collection string, id string, vector []float32, payload map[string]interface{}) (string, error) {
	url := fmt.Sprintf("%s/collections/%s/points?wait=true", r.baseURL, collection)

	// Qdrant expects this exact format
	body := map[string]interface{}{
		"points": []map[string]interface{}{
			{
				"id":      id, // UUID string is fine
				"vector":  vector,
				"payload": payload,
			},
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal body: %w", err)
	}

	log.Printf("📤 Sending to Qdrant: URL=%s, ID=%s, VectorLen=%d", url, id, len(vector))

	resp, err := r.client.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("qdrant request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		// Read error details
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Qdrant error response: Status=%d, Body=%s", resp.StatusCode, string(body))
		return "", fmt.Errorf("qdrant responded %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("✅ Successfully stored vector ID=%s", id)
	return id, nil
}

// EnsureCollection creates the collection if it doesn't exist
func (r *QdrantRepo) EnsureCollection(collection string, dim int) error {
	url := fmt.Sprintf("%s/collections/%s", r.baseURL, collection)

	// Check if collection exists
	resp, err := r.client.Get(url)
	if err == nil && resp.StatusCode == 200 {
		resp.Body.Close()
		log.Printf("✅ Collection '%s' already exists", collection)
		return nil
	}
	if resp != nil {
		resp.Body.Close()
	}

	// Create collection
	log.Printf("📦 Creating collection '%s' with dimension %d", collection, dim)

	body := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     dim,
			"distance": "Cosine",
		},
	}
	b, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp2, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp2.Body)
		log.Printf("❌ Failed to create collection: Status=%d, Body=%s", resp2.StatusCode, string(bodyBytes))
		return fmt.Errorf("create collection returned %d: %s", resp2.StatusCode, string(bodyBytes))
	}

	log.Printf("✅ Collection '%s' created successfully", collection)
	return nil
}

// Search performs similarity search
func (r *QdrantRepo) Search(collection string, vector []float32, limit int) ([]SearchResult, error) {
	url := fmt.Sprintf("%s/collections/%s/points/search", r.baseURL, collection)
	body := map[string]interface{}{
		"vector":       vector,
		"limit":        limit,
		"with_payload": true,
	}
	b, _ := json.Marshal(body)

	resp, err := r.client.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("qdrant search responded %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Result []SearchResult `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Result, nil
}

type SearchResult struct {
	ID      string                 `json:"id"`
	Score   float64                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
}
