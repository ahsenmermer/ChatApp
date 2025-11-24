package models

type Chunk struct {
	ChunkID     string                 `json:"chunk_id"`
	Text        string                 `json:"text"`
	Page        int                    `json:"page,omitempty"`
	StartOffset int                    `json:"start_offset,omitempty"`
	EndOffset   int                    `json:"end_offset,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type OCRProcessedEvent struct {
	Event       string  `json:"event"`
	FileID      string  `json:"file_id"`
	FileName    string  `json:"file_name"`
	ContentType string  `json:"content_type"`
	Chunks      []Chunk `json:"chunks"`
	Language    string  `json:"language,omitempty"`
	Timestamp   string  `json:"timestamp"`
}

type StoredChunkInfo struct {
	ChunkID  string `json:"chunk_id"`
	VectorID string `json:"vector_id"`
	Page     int    `json:"page"`
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
