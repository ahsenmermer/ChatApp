package services

import (
	"chat_service/internal/models"
	"sync"
)

type FileTracker struct {
	files map[string]*models.FileStatus
	mu    sync.RWMutex
}

func NewFileTracker() *FileTracker {
	return &FileTracker{
		files: make(map[string]*models.FileStatus),
	}
}

func (f *FileTracker) UpdateStatus(fileID, fileName, status string, totalChunks int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.files[fileID]; !exists {
		f.files[fileID] = &models.FileStatus{
			FileID:   fileID,
			FileName: fileName,
		}
	}

	f.files[fileID].Status = status
	f.files[fileID].TotalChunks = totalChunks
}

func (f *FileTracker) GetStatus(fileID string) *models.FileStatus {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.files[fileID]
}

func (f *FileTracker) IsReady(fileID string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if status, exists := f.files[fileID]; exists {
		return status.Status == "ready"
	}
	return false
}
