package services

import (
	"context"
	"fmt"
	"log"
	"strings"

	"chat_service/internal/repository"
)

type RAGService struct {
	qdrantClient *repository.QdrantClient
	xenovaClient *XenovaClient
}

func NewRAGService(qdrantClient *repository.QdrantClient, xenovaClient *XenovaClient) *RAGService {
	return &RAGService{
		qdrantClient: qdrantClient,
		xenovaClient: xenovaClient,
	}
}

type RelevantChunk struct {
	ChunkID  string
	Text     string
	Score    float64
	Page     int
	FileName string
}

func (r *RAGService) SearchRelevantChunks(ctx context.Context, query string, fileID string, limit int) ([]RelevantChunk, error) {
	log.Printf("🔄 Embedding query: %s", truncateString(query, 50))

	queryVector, err := r.xenovaClient.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	log.Printf("✅ Query embedded, dimension: %d", len(queryVector))

	filter := map[string]interface{}{
		"must": []map[string]interface{}{
			{
				"key": "file_id",
				"match": map[string]interface{}{
					"value": fileID,
				},
			},
		},
	}

	log.Printf("🔍 Searching in Qdrant for file_id: %s (limit: %d)", fileID, limit)

	results, err := r.qdrantClient.Search("documents", queryVector, limit, filter)
	if err != nil {
		return nil, fmt.Errorf("qdrant search failed: %w", err)
	}

	log.Printf("📊 Found %d results from Qdrant", len(results))

	var chunks []RelevantChunk
	for i, res := range results {
		text, _ := res.Payload["text"].(string)
		fileName, _ := res.Payload["file_name"].(string)

		page := 0
		if p, ok := res.Payload["page"].(float64); ok {
			page = int(p)
		}

		chunks = append(chunks, RelevantChunk{
			ChunkID:  res.ID,
			Text:     text,
			Score:    res.Score,
			Page:     page,
			FileName: fileName,
		})

		log.Printf("  Result %d: Page %d, Score %.3f, Text length: %d",
			i+1, page, res.Score, len(text))
	}

	return chunks, nil
}

func (r *RAGService) BuildContext(chunks []RelevantChunk) string {
	if len(chunks) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("# Belgeden İlgili Bölümler:\n\n")

	for i, chunk := range chunks {
		builder.WriteString(fmt.Sprintf("## Bölüm %d (Sayfa %d, Benzerlik: %.1f%%)\n",
			i+1, chunk.Page, chunk.Score*100))
		builder.WriteString(chunk.Text)
		builder.WriteString("\n\n")
	}

	contextStr := builder.String()
	log.Printf("📝 Built RAG context: %d characters from %d chunks", len(contextStr), len(chunks))

	return contextStr
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
