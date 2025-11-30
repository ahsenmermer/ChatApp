package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"embedding_service/internal/repository"
)

type SearchService struct {
	qrepo  *repository.QdrantRepo
	xenova *XenovaClient
}

func NewSearchService(qrepo *repository.QdrantRepo, xenova *XenovaClient) *SearchService {
	return &SearchService{
		qrepo:  qrepo,
		xenova: xenova,
	}
}

type SearchRequest struct {
	Query    string  `json:"query"`
	Limit    int     `json:"limit,omitempty"`
	FileID   string  `json:"file_id,omitempty"`   // Belirli bir dosyada ara
	MinScore float64 `json:"min_score,omitempty"` // Minimum benzerlik skoru
}

type SearchResponse struct {
	Results []SearchResultItem `json:"results"`
	Query   string             `json:"query"`
	Total   int                `json:"total"`
}

type SearchResultItem struct {
	ChunkID  string                 `json:"chunk_id"`
	Text     string                 `json:"text"`
	Score    float64                `json:"score"`
	FileName string                 `json:"file_name"`
	FileID   string                 `json:"file_id"`
	Page     int                    `json:"page,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// HandleSearch handles semantic search requests
func (s *SearchService) HandleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "query field is required", http.StatusBadRequest)
		return
	}

	// Default limit
	if req.Limit <= 0 {
		req.Limit = 5
	}
	if req.Limit > 50 {
		req.Limit = 50 // Max limit
	}

	log.Printf("🔍 Search request: query='%s', limit=%d, file_id='%s'", req.Query, req.Limit, req.FileID)

	// 1. Query'yi embedding'e çevir
	queryVec, err := s.xenova.Embed(req.Query)
	if err != nil {
		log.Printf("Failed to embed query: %v", err)
		http.Error(w, fmt.Sprintf("Embedding failed: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Query embedded, dimension: %d", len(queryVec))

	// 2. Qdrant'ta ara
	results, err := s.qrepo.Search("documents", queryVec, req.Limit*2) // 2x al, filtreleyeceğiz
	if err != nil {
		log.Printf("Qdrant search failed: %v", err)
		http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("📊 Found %d results from Qdrant", len(results))

	// 3. Sonuçları formatla ve filtrele
	var items []SearchResultItem
	for _, r := range results {
		// Min score filtresi
		if req.MinScore > 0 && r.Score < req.MinScore {
			continue
		}

		// FileID filtresi
		if req.FileID != "" {
			if fileID, ok := r.Payload["file_id"].(string); !ok || fileID != req.FileID {
				continue
			}
		}

		item := SearchResultItem{
			ChunkID:  r.ID,
			Score:    r.Score,
			Metadata: r.Payload,
		}

		// Payload'dan alanları çıkar
		if text, ok := r.Payload["text"].(string); ok {
			item.Text = text
		}
		if fileName, ok := r.Payload["file_name"].(string); ok {
			item.FileName = fileName
		}
		if fileID, ok := r.Payload["file_id"].(string); ok {
			item.FileID = fileID
		}
		if page, ok := r.Payload["page"].(float64); ok {
			item.Page = int(page)
		}

		items = append(items, item)

		// Limit'e ulaştıysak dur
		if len(items) >= req.Limit {
			break
		}
	}

	log.Printf("✅ Returning %d filtered results", len(items))

	// 4. Response döndür
	response := SearchResponse{
		Results: items,
		Query:   req.Query,
		Total:   len(items),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *SearchService) HandleSearchByFileID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fileID := r.URL.Query().Get("file_id")
	if fileID == "" {
		http.Error(w, "file_id parameter required", http.StatusBadRequest)
		return
	}

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// ✅ FileID'yi ekle ve direkt arama yap
	req.FileID = fileID

	if req.Query == "" {
		http.Error(w, "query field is required", http.StatusBadRequest)
		return
	}

	// Default limit
	if req.Limit <= 0 {
		req.Limit = 5
	}
	if req.Limit > 50 {
		req.Limit = 50
	}

	log.Printf("🔍 File search: query='%s', file_id='%s', limit=%d", req.Query, req.FileID, req.Limit)

	// Query'yi embed et
	queryVec, err := s.xenova.Embed(req.Query)
	if err != nil {
		log.Printf("Failed to embed query: %v", err)
		http.Error(w, fmt.Sprintf("Embedding failed: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Query embedded, dimension: %d", len(queryVec))

	// Qdrant'ta ara
	results, err := s.qrepo.Search("documents", queryVec, req.Limit*2)
	if err != nil {
		log.Printf("Qdrant search failed: %v", err)
		http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("📊 Found %d results from Qdrant", len(results))

	// Sonuçları filtrele
	var items []SearchResultItem
	for _, r := range results {
		// Min score filtresi
		if req.MinScore > 0 && r.Score < req.MinScore {
			continue
		}

		// FileID filtresi (ZORUNLU)
		if fileID, ok := r.Payload["file_id"].(string); !ok || fileID != req.FileID {
			continue
		}

		item := SearchResultItem{
			ChunkID:  r.ID,
			Score:    r.Score,
			Metadata: r.Payload,
		}

		if text, ok := r.Payload["text"].(string); ok {
			item.Text = text
		}
		if fileName, ok := r.Payload["file_name"].(string); ok {
			item.FileName = fileName
		}
		if fileIDVal, ok := r.Payload["file_id"].(string); ok {
			item.FileID = fileIDVal
		}
		if page, ok := r.Payload["page"].(float64); ok {
			item.Page = int(page)
		}

		items = append(items, item)

		if len(items) >= req.Limit {
			break
		}
	}

	log.Printf("✅ Returning %d filtered results for file %s", len(items), req.FileID)

	response := SearchResponse{
		Results: items,
		Query:   req.Query,
		Total:   len(items),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
