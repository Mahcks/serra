package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

type csrfTokenStore struct {
	mu     sync.RWMutex
	tokens map[string]time.Time
}

func newCSRFStore() *csrfTokenStore {
	store := &csrfTokenStore{
		tokens: make(map[string]time.Time),
	}

	// Start cleanup goroutine
	go store.cleanup()

	return store
}

func (cs *csrfTokenStore) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	token := hex.EncodeToString(bytes)

	cs.mu.Lock()
	cs.tokens[token] = time.Now().Add(1 * time.Hour) // 1 hour expiry
	cs.mu.Unlock()

	return token, nil
}

func (cs *csrfTokenStore) validateToken(token string) bool {
	cs.mu.RLock()
	expiry, exists := cs.tokens[token]
	cs.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(expiry) {
		// Remove expired token
		cs.mu.Lock()
		delete(cs.tokens, token)
		cs.mu.Unlock()
		return false
	}

	// Remove used token (single use)
	cs.mu.Lock()
	delete(cs.tokens, token)
	cs.mu.Unlock()

	return true
}

func (cs *csrfTokenStore) cleanup() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cs.mu.Lock()
		now := time.Now()
		for token, expiry := range cs.tokens {
			if now.After(expiry) {
				delete(cs.tokens, token)
			}
		}
		cs.mu.Unlock()
	}
}

var csrfStore = newCSRFStore()

// CSRFProtection provides CSRF protection for state-changing operations
func CSRFProtection() fiber.Handler {
	return func(c *fiber.Ctx) error {
		method := c.Method()

		// Only protect state-changing methods
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			return c.Next()
		}

		// Get CSRF token from header
		token := c.Get("X-CSRF-Token")
		if token == "" {
			return apiErrors.ErrBadRequest().SetDetail("CSRF token required")
		}

		// Validate token
		if !csrfStore.validateToken(token) {
			return apiErrors.ErrForbidden().SetDetail("Invalid or expired CSRF token")
		}

		return c.Next()
	}
}

// GenerateCSRFToken generates a new CSRF token
func GenerateCSRFToken() (string, error) {
	return csrfStore.generateToken()
}
