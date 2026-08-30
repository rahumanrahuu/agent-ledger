package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter implements token bucket algorithm for rate limiting
type RateLimiter struct {
	mu              sync.RWMutex
	buckets         map[string]*tokenBucket
	requestsPerSec  float64
	burstSize       int
	windowSize      time.Duration
	cleanupInterval time.Duration
	stopCleanup     chan bool
}

// tokenBucket represents a rate limit bucket for a client
type tokenBucket struct {
	tokens      float64
	lastRefill  time.Time
	totalUsed   int64
	lastReset   time.Time
	blocked     bool
	blockReason string
	blockUntil  time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSec float64, burstSize int) *RateLimiter {
	rl := &RateLimiter{
		buckets:         make(map[string]*tokenBucket),
		requestsPerSec:  requestsPerSec,
		burstSize:       burstSize,
		windowSize:      time.Minute,
		cleanupInterval: 10 * time.Minute,
		stopCleanup:     make(chan bool),
	}

	// Start cleanup goroutine
	go rl.cleanupExpiredBuckets()

	return rl
}

// Allow checks if a request is allowed for the given client
func (rl *RateLimiter) Allow(clientID string) (bool, *RateLimitInfo) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[clientID]
	if !exists {
		bucket = &tokenBucket{
			tokens:     float64(rl.burstSize),
			lastRefill: time.Now(),
			lastReset:  time.Now(),
		}
		rl.buckets[clientID] = bucket
	}

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	bucket.tokens += elapsed * rl.requestsPerSec
	if bucket.tokens > float64(rl.burstSize) {
		bucket.tokens = float64(rl.burstSize)
	}
	bucket.lastRefill = now

	// Check if blocked
	if bucket.blocked && now.Before(bucket.blockUntil) {
		info := &RateLimitInfo{
			Allowed:      false,
			ClientID:     clientID,
			TokensUsed:   1,
			TokensRemaining: 0,
			ResetTime:    bucket.blockUntil,
			Blocked:      true,
			BlockReason:  bucket.blockReason,
		}
		return false, info
	}

	// Reset block status if time has passed
	if bucket.blocked && now.After(bucket.blockUntil) {
		bucket.blocked = false
		bucket.tokens = float64(rl.burstSize)
	}

	// Check if request is allowed
	if bucket.tokens >= 1.0 {
		bucket.tokens--
		bucket.totalUsed++
		info := &RateLimitInfo{
			Allowed:         true,
			ClientID:        clientID,
			TokensUsed:      1,
			TokensRemaining: int64(bucket.tokens),
			ResetTime:       bucket.lastReset.Add(rl.windowSize),
			Blocked:         false,
		}
		return true, info
	}

	// Request denied
	info := &RateLimitInfo{
		Allowed:         false,
		ClientID:        clientID,
		TokensUsed:      0,
		TokensRemaining: int64(bucket.tokens),
		ResetTime:       bucket.lastReset.Add(rl.windowSize),
		Blocked:         false,
	}
	return false, info
}

// Block blocks a client for a specific duration
func (rl *RateLimiter) Block(clientID string, duration time.Duration, reason string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[clientID]
	if !exists {
		bucket = &tokenBucket{}
		rl.buckets[clientID] = bucket
	}

	bucket.blocked = true
	bucket.blockReason = reason
	bucket.blockUntil = time.Now().Add(duration)
	bucket.tokens = 0

	return nil
}

// Unblock removes the block from a client
func (rl *RateLimiter) Unblock(clientID string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[clientID]
	if !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	bucket.blocked = false
	bucket.tokens = float64(rl.burstSize)

	return nil
}

// RateLimitInfo contains rate limit information
type RateLimitInfo struct {
	Allowed         bool
	ClientID        string
	TokensUsed      int64
	TokensRemaining int64
	ResetTime       time.Time
	Blocked         bool
	BlockReason     string
}

// GetStats returns statistics for a client
func (rl *RateLimiter) GetStats(clientID string) *RateLimitStats {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	bucket, exists := rl.buckets[clientID]
	if !exists {
		return &RateLimitStats{
			ClientID: clientID,
			Found:    false,
		}
	}

	windowStart := bucket.lastReset
	windowEnd := windowStart.Add(rl.windowSize)
	totalRequests := bucket.totalUsed
	if windowEnd.Before(time.Now()) {
		totalRequests = 0
		windowStart = time.Now()
		windowEnd = windowStart.Add(rl.windowSize)
	}

	return &RateLimitStats{
		ClientID:        clientID,
		Found:           true,
		TokensRemaining: int64(bucket.tokens),
		MaxTokens:       int64(rl.burstSize),
		TotalRequests:   totalRequests,
		WindowStart:     windowStart,
		WindowEnd:       windowEnd,
		Blocked:         bucket.blocked,
		BlockReason:     bucket.blockReason,
		BlockUntil:      bucket.blockUntil,
	}
}

// RateLimitStats contains rate limiting statistics
type RateLimitStats struct {
	ClientID        string
	Found           bool
	TokensRemaining int64
	MaxTokens       int64
	TotalRequests   int64
	WindowStart     time.Time
	WindowEnd       time.Time
	Blocked         bool
	BlockReason     string
	BlockUntil      time.Time
}

// cleanupExpiredBuckets removes old buckets periodically
func (rl *RateLimiter) cleanupExpiredBuckets() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.removeExpiredBuckets()
		case <-rl.stopCleanup:
			return
		}
	}
}

// removeExpiredBuckets removes buckets older than 24 hours
func (rl *RateLimiter) removeExpiredBuckets() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour)
	for clientID, bucket := range rl.buckets {
		if bucket.lastRefill.Before(cutoff) {
			delete(rl.buckets, clientID)
		}
	}
}

// Close stops the rate limiter
func (rl *RateLimiter) Close() {
	rl.stopCleanup <- true
}

// ListClients returns all active clients
func (rl *RateLimiter) ListClients() []string {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	clients := make([]string, 0, len(rl.buckets))
	for clientID := range rl.buckets {
		clients = append(clients, clientID)
	}
	return clients
}
