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

type QdrantRepo struct {
	baseURL string
	client  *http.Client
}

func NewQdrantRepo(baseURL string) *QdrantRepo {
	return &QdrantRepo{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// UpsertVector uploads a single vector to Qdrant
func (r *QdrantRepo) UpsertVector(collection string, id string, vector []float32, payload map[string]interface{}) (string, error) {
	// ✅ Qdrant 1.x için doğru endpoint (/upsert YOK)
	url := fmt.Sprintf("%s/collections/%s/points?wait=true", r.baseURL, collection)

	// Payload'u temizle - sadece JSON-safe değerleri tut
	cleanedPayload := make(map[string]interface{})
	for k, v := range payload {
		switch val := v.(type) {
		case string, bool, int64, float64:
			cleanedPayload[k] = v
		case int:
			cleanedPayload[k] = int64(val)
		case int32:
			cleanedPayload[k] = int64(val)
		case float32:
			cleanedPayload[k] = float64(val)
		case nil:
			continue
		default:
			cleanedPayload[k] = fmt.Sprintf("%v", v)
		}
	}

	// ✅ vector'ü float64'e çevir
	vectorFloat64 := make([]float64, len(vector))
	for i, v := range vector {
		vectorFloat64[i] = float64(v)
	}

	// ✅ Batch upsert format (Qdrant 1.x için gerekli)
	body := map[string]interface{}{
		"batch": map[string]interface{}{
			"ids":      []string{id},
			"vectors":  [][]float64{vectorFloat64},
			"payloads": []map[string]interface{}{cleanedPayload},
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal body: %w", err)
	}

	// ✅ PUT method kullan (POST yerine)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("qdrant request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Qdrant Error: Status=%d, URL=%s", resp.StatusCode, url)
		log.Printf("   Request body (first 500 chars): %s", truncate(string(b), 500))
		log.Printf("   Response: %s", string(bodyBytes))
		return "", fmt.Errorf("qdrant responded %d: %s", resp.StatusCode, string(bodyBytes))
	}

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

	// ✅ vector'ü float64'e çevir
	vectorFloat64 := make([]float64, len(vector))
	for i, v := range vector {
		vectorFloat64[i] = float64(v)
	}

	body := map[string]interface{}{
		"vector":       vectorFloat64,
		"limit":        limit,
		"with_payload": true,
		"with_vector":  false,
	}

	b, _ := json.Marshal(body)
	resp, err := r.client.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Qdrant search error: Status=%d, Response=%s", resp.StatusCode, string(bodyBytes))
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

type SearchResult struct {
	ID      string                 `json:"id"`
	Score   float64                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
}

// Helper function
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
