package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ocr_service/internal/events"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadHandler struct {
	Producer *events.Producer
}

func NewUploadHandler(p *events.Producer) *UploadHandler {
	return &UploadHandler{Producer: p}
}

func (h *UploadHandler) HandleUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("❌ Upload error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "file field is required"})
		return
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	validExts := []string{".pdf", ".png", ".jpg", ".jpeg", ".tiff", ".tif", ".bmp"}
	isValid := false
	for _, validExt := range validExts {
		if ext == validExt {
			isValid = true
			break
		}
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid file type. Supported: PDF, PNG, JPG, JPEG, TIFF, BMP",
		})
		return
	}

	// Validate file size (max 50MB)
	if file.Size > 50*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file size exceeds 50MB limit",
		})
		return
	}

	// Create uploads directory if not exists
	uploadDir := "/tmp/uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			log.Printf("❌ Failed to create upload directory: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
			return
		}
	}

	fileID := uuid.New().String()
	dst := filepath.Join(uploadDir, fileID+"_"+file.Filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		log.Printf("❌ Failed to save file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	log.Printf("📁 File saved: %s (ID: %s, Size: %d bytes)", file.Filename, fileID, file.Size)

	// Set initial status
	SetFileStatus(fileID, "uploaded", 0, "File uploaded successfully, queued for processing")

	// Publish FILE_UPLOADED event
	event := map[string]interface{}{
		"event":        "FILE_UPLOADED",
		"file_id":      fileID,
		"file_name":    file.Filename,
		"file_path":    dst,
		"content_type": file.Header.Get("Content-Type"),
		"file_size":    file.Size,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	}
	b, _ := json.Marshal(event)

	if err := h.Producer.Publish("file_uploaded", b); err != nil {
		log.Printf("❌ Failed to publish event: %v", err)
		SetFileStatus(fileID, "failed", 0, "Failed to queue file for processing")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue file"})
		return
	}

	log.Printf("✅ Published FILE_UPLOADED event for file %s", fileID)

	c.JSON(http.StatusAccepted, gin.H{
		"file_id":   fileID,
		"file_name": file.Filename,
		"status":    "accepted",
		"message":   "File queued for processing",
	})
}
