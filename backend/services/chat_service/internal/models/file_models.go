package models

type FileStatus struct {
	FileID         string `json:"file_id"`
	FileName       string `json:"file_name"`
	Status         string `json:"status"` // "processing", "ready", "failed"
	TotalChunks    int    `json:"total_chunks"`
	UploadedAt     string `json:"uploaded_at"`
	UserID         string `json:"user_id"`         // ✅ YENİ
	ConversationID string `json:"conversation_id"` // ✅ YENİ
}

type EmbeddingStoredEvent struct {
	Event       string            `json:"event"`
	FileID      string            `json:"file_id"`
	FileName    string            `json:"file_name"`
	ContentType string            `json:"content_type"`
	Chunks      []StoredChunkInfo `json:"chunks"`
	TotalChunks int               `json:"total_chunks"`
	Timestamp   string            `json:"timestamp"`
}

type StoredChunkInfo struct {
	ChunkID  string `json:"chunk_id"`
	VectorID string `json:"vector_id"`
	Page     int    `json:"page"`
}
