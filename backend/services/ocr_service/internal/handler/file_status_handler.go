package handler

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type FileStatus struct {
	FileID      string `json:"file_id"`
	Status      string `json:"status"` // processing, completed, failed
	TotalChunks int    `json:"total_chunks"`
	Message     string `json:"message,omitempty"`
}

var (
	fileStatuses = make(map[string]*FileStatus)
	statusMutex  sync.RWMutex
)

// SetFileStatus updates file processing status
func SetFileStatus(fileID, status string, totalChunks int, message string) {
	statusMutex.Lock()
	defer statusMutex.Unlock()

	fileStatuses[fileID] = &FileStatus{
		FileID:      fileID,
		Status:      status,
		TotalChunks: totalChunks,
		Message:     message,
	}
}

// GetFileStatus returns current file status
func GetFileStatus(fileID string) (*FileStatus, bool) {
	statusMutex.RLock()
	defer statusMutex.RUnlock()

	status, exists := fileStatuses[fileID]
	return status, exists
}

// HandleFileStatus is the HTTP handler for checking file status
func HandleFileStatus(c *gin.Context) {
	fileID := c.Param("file_id")

	status, exists := GetFileStatus(fileID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "file not found or not yet processed",
		})
		return
	}

	c.JSON(http.StatusOK, status)
}
