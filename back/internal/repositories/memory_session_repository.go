package repositories

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mon-gene/back/internal/models"
)

type memorySessionRepository struct {
	sessions map[string]*models.Session
	mutex    sync.RWMutex
}

func NewMemorySessionRepository() SessionRepository {
	return &memorySessionRepository{
		sessions: make(map[string]*models.Session),
		mutex:    sync.RWMutex{},
	}
}

func (r *memorySessionRepository) Create(ctx context.Context, session *models.Session) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.sessions[session.ID] = session
	return nil
}

func (r *memorySessionRepository) GetByToken(ctx context.Context, token string) (*models.Session, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	session, exists := r.sessions[token]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}
	
	return session, nil
}

func (r *memorySessionRepository) Delete(ctx context.Context, token string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	delete(r.sessions, token)
	return nil
}

func (r *memorySessionRepository) DeleteExpired(ctx context.Context) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	now := time.Now()
	for token, session := range r.sessions {
		if now.After(session.ExpiresAt) {
			delete(r.sessions, token)
		}
	}
	
	return nil
}
