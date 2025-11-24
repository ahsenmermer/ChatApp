package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}

	if _, err := os.Stat("/tmp/uploads"); os.IsNotExist(err) {
		os.MkdirAll("/tmp/uploads", 0755)
	}

	fileID := uuid.New().String()
	dst := filepath.Join("/tmp/uploads", fileID+"_"+file.Filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
		return
	}

	// Set initial status
	SetFileStatus(fileID, "uploaded", 0, "File uploaded, waiting for processing")

	event := map[string]interface{}{
		"event":        "FILE_UPLOADED",
		"file_id":      fileID,
		"file_name":    file.Filename,
		"file_path":    dst,
		"content_type": file.Header.Get("Content-Type"),
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	}
	b, _ := json.Marshal(event)

	if err := h.Producer.Publish("file_uploaded", b); err != nil {
		log.Printf("publish failed: %v", err)
		SetFileStatus(fileID, "failed", 0, "Failed to publish upload event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "publish failed"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"file_id": fileID,
		"status":  "accepted",
		"message": "File queued for processing",
	})
}
