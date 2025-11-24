package models

type FileUploadedEvent struct {
	Event       string `json:"event"`
	FileID      string `json:"file_id"`
	FileName    string `json:"file_name"`
	FilePath    string `json:"file_path"`
	ContentType string `json:"content_type"`
	UploaderID  string `json:"uploader_id,omitempty"`
	Timestamp   string `json:"timestamp"`
}

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
