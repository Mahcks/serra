package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

type rateLimiter struct {
	mu         sync.Mutex
	requests   map[string][]time.Time
	maxReqs    int
	windowSize time.Duration
}

func newRateLimiter(maxReqs int, windowSize time.Duration) *rateLimiter {
	return &rateLimiter{
		requests:   make(map[string][]time.Time),
		maxReqs:    maxReqs,
		windowSize: windowSize,
	}
}

func (rl *rateLimiter) isAllowed(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.windowSize)

	// Get existing requests for this key
	requests := rl.requests[key]
	
	// Filter out old requests
	var validRequests []time.Time
	for _, reqTime := range requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if we're under the limit
	if len(validRequests) >= rl.maxReqs {
		return false
	}

	// Add this request
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests

	return true
}

// Clean up old entries periodically
func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.windowSize)

	for key, requests := range rl.requests {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}
		
		if len(validRequests) == 0 {
			delete(rl.requests, key)
		} else {
			rl.requests[key] = validRequests
		}
	}
}

var (
	// 5 requests per minute for invitation acceptance
	inviteRateLimiter = newRateLimiter(5, time.Minute)
)

// Start cleanup goroutine
func init() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				inviteRateLimiter.cleanup()
			}
		}
	}()
}

// RateLimitInvitations limits invitation-related requests by IP
func RateLimitInvitations() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientIP := c.IP()
		
		if !inviteRateLimiter.isAllowed(clientIP) {
			return apiErrors.ErrTooManyRequests().SetDetail("Too many invitation requests. Please try again later.")
		}
		
		return c.Next()
	}
}