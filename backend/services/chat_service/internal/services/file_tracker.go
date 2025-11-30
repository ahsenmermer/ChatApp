package services

import (
	"chat_service/internal/models"
	"log"
	"sync"
	"time"
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
			FileID:     fileID,
			FileName:   fileName,
			UploadedAt: time.Now().Format(time.RFC3339),
		}
	}

	f.files[fileID].Status = status
	f.files[fileID].TotalChunks = totalChunks

	log.Printf("📝 FileTracker updated: %s -> status=%s, chunks=%d",
		fileID, status, totalChunks)
}

// ✅ DÜZELTME: Dosya yoksa oluştur!
func (f *FileTracker) SetUserInfo(fileID, userID, conversationID string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// ✅ Eğer dosya yoksa önce oluştur
	if _, exists := f.files[fileID]; !exists {
		f.files[fileID] = &models.FileStatus{
			FileID:     fileID,
			UploadedAt: time.Now().Format(time.RFC3339),
			Status:     "pending", // Geçici status
		}
		log.Printf("📝 FileTracker: Created entry for %s", fileID)
	}

	// User bilgilerini ekle
	f.files[fileID].UserID = userID
	f.files[fileID].ConversationID = conversationID

	log.Printf("📝 FileTracker: Added user info for %s (user=%s, conv=%s)",
		fileID, userID, conversationID)
}

func (f *FileTracker) GetFileInfo(fileID string) *models.FileStatus {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.files[fileID]
}

func (f *FileTracker) GetStatus(fileID string) *models.FileStatus {
	f.mu.RLock()
	defer f.mu.RUnlock()

	status := f.files[fileID]

	if status == nil {
		log.Printf("⚠️ FileTracker: File not found: %s", fileID)
	} else {
		log.Printf("📊 FileTracker: File %s -> status=%s, chunks=%d",
			fileID, status.Status, status.TotalChunks)
	}

	return status
}

func (f *FileTracker) IsReady(fileID string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if status, exists := f.files[fileID]; exists {
		isReady := status.Status == "ready" && status.TotalChunks > 0
		log.Printf("🔍 FileTracker.IsReady(%s): exists=%v, status=%s, chunks=%d -> %v",
			fileID, exists, status.Status, status.TotalChunks, isReady)
		return isReady
	}

	log.Printf("🔍 FileTracker.IsReady(%s): NOT FOUND in tracker", fileID)
	return false
}
