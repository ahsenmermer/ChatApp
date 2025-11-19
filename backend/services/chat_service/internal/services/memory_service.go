package services

import (
	"sync"
)

// LangChain tarzı bellek yapısı (user_id -> geçmiş mesajlar)
type MemoryService struct {
	memory map[string][]string
	mu     sync.RWMutex
	maxCtx int // en fazla kaç mesaj tutulacak
}

func NewMemoryService(maxCtx int) *MemoryService {
	return &MemoryService{
		memory: make(map[string][]string),
		maxCtx: maxCtx,
	}
}

// Geçmişe mesaj ekler
func (m *MemoryService) AddMessage(userID, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.memory[userID] = append(m.memory[userID], message)
	if len(m.memory[userID]) > m.maxCtx {
		m.memory[userID] = m.memory[userID][len(m.memory[userID])-m.maxCtx:]
	}
}

// Geçmişi döner (LangChain prompt context gibi)
func (m *MemoryService) GetContext(userID string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	context := ""
	for _, msg := range m.memory[userID] {
		context += msg + "\n"
	}
	return context
}
