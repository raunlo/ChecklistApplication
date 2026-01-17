package service

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

type ISessionService interface {
	GenerateClientId(userId string) (string, error)
	ValidateClientId(clientId string, userId string) bool
	CleanupExpired()
}

type session struct {
	userId    string
	expiresAt time.Time
}

type sessionServiceImpl struct {
	sessions map[string]*session // clientId -> session
	mu       sync.RWMutex
	ttl      time.Duration
}

func NewSessionService(ttl time.Duration) ISessionService {
	s := &sessionServiceImpl{
		sessions: make(map[string]*session),
		ttl:      ttl,
	}

	// Start background cleanup goroutine
	go s.cleanupLoop()

	return s
}

// GenerateClientId creates a new secure client ID for the user
func (s *sessionServiceImpl) GenerateClientId(userId string) (string, error) {
	// Generate 32 bytes of random data
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	clientId := base64.URLEncoding.EncodeToString(bytes)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[clientId] = &session{
		userId:    userId,
		expiresAt: time.Now().Add(s.ttl),
	}

	return clientId, nil
}

// ValidateClientId checks if the client ID is valid for the given user
func (s *sessionServiceImpl) ValidateClientId(clientId string, userId string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sess, exists := s.sessions[clientId]
	if !exists {
		return false
	}

	// Check if expired
	if time.Now().After(sess.expiresAt) {
		return false
	}

	// Check if belongs to user
	return sess.userId == userId
}

// CleanupExpired removes expired sessions (public for testing)
func (s *sessionServiceImpl) CleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for clientId, sess := range s.sessions {
		if now.After(sess.expiresAt) {
			delete(s.sessions, clientId)
		}
	}
}

// cleanupLoop runs periodic cleanup every hour
func (s *sessionServiceImpl) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.CleanupExpired()
	}
}
