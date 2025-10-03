package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter provides rate limiting for API requests
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex

	// Default limits
	requestsPerSecond int
	burst             int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond, burst int) *RateLimiter {
	return &RateLimiter{
		limiters:          make(map[string]*rate.Limiter),
		requestsPerSecond: requestsPerSecond,
		burst:             burst,
	}
}

// GetLimiter returns or creates a rate limiter for a key
func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(rl.requestsPerSecond), rl.burst)
		rl.limiters[key] = limiter
	}

	return limiter
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(key string) bool {
	limiter := rl.GetLimiter(key)
	return limiter.Allow()
}

// Wait waits until a request is allowed or context is cancelled
func (rl *RateLimiter) Wait(ctx context.Context, key string) error {
	limiter := rl.GetLimiter(key)
	return limiter.Wait(ctx)
}

// Reserve reserves a token and returns a reservation
func (rl *RateLimiter) Reserve(key string) *rate.Reservation {
	limiter := rl.GetLimiter(key)
	return limiter.Reserve()
}

// MultiLimiter handles multiple rate limit tiers
type MultiLimiter struct {
	limiters []*tierLimiter
	mu       sync.RWMutex
}

type tierLimiter struct {
	name    string
	limiter *rate.Limiter
	window  time.Duration
}

// NewMultiLimiter creates a multi-tier rate limiter
func NewMultiLimiter() *MultiLimiter {
	return &MultiLimiter{
		limiters: make([]*tierLimiter, 0),
	}
}

// AddTier adds a rate limit tier
func (ml *MultiLimiter) AddTier(name string, limit int, window time.Duration) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	// Calculate requests per second for this tier
	rps := float64(limit) / window.Seconds()

	ml.limiters = append(ml.limiters, &tierLimiter{
		name:    name,
		limiter: rate.NewLimiter(rate.Limit(rps), limit),
		window:  window,
	})
}

// Allow checks if request is allowed across all tiers
func (ml *MultiLimiter) Allow() bool {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	for _, tier := range ml.limiters {
		if !tier.limiter.Allow() {
			return false
		}
	}
	return true
}

// Wait waits for all tiers to allow the request
func (ml *MultiLimiter) Wait(ctx context.Context) error {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	for _, tier := range ml.limiters {
		if err := tier.limiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limit tier %s: %w", tier.name, err)
		}
	}
	return nil
}

// TokenBucket implements a simple token bucket rate limiter
type TokenBucket struct {
	tokens    float64
	capacity  float64
	refillRate float64
	lastRefill time.Time
	mu        sync.Mutex
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Take attempts to take n tokens from the bucket
func (tb *TokenBucket) Take(n float64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens >= n {
		tb.tokens -= n
		return true
	}
	return false
}

// Wait waits until n tokens are available
func (tb *TokenBucket) Wait(ctx context.Context, n float64) error {
	for {
		if tb.Take(n) {
			return nil
		}

		// Calculate wait time
		tb.mu.Lock()
		needed := n - tb.tokens
		waitTime := time.Duration(needed/tb.refillRate) * time.Second
		tb.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Try again
		}
	}
}

// refill refills the bucket based on elapsed time
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	newTokens := elapsed * tb.refillRate
	tb.tokens = min(tb.capacity, tb.tokens+newTokens)
	tb.lastRefill = now
}

// Available returns the number of available tokens
func (tb *TokenBucket) Available() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()
	return tb.tokens
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// WeightedRateLimiter applies different limits based on weight
type WeightedRateLimiter struct {
	limiters map[int]*rate.Limiter
	mu       sync.RWMutex
	baseRate float64
}

// NewWeightedRateLimiter creates a weighted rate limiter
func NewWeightedRateLimiter(baseRate float64) *WeightedRateLimiter {
	return &WeightedRateLimiter{
		limiters: make(map[int]*rate.Limiter),
		baseRate: baseRate,
	}
}

// Allow checks if a request with given weight is allowed
func (wrl *WeightedRateLimiter) Allow(weight int) bool {
	wrl.mu.Lock()
	limiter, exists := wrl.limiters[weight]
	if !exists {
		// Create limiter with adjusted rate based on weight
		adjustedRate := wrl.baseRate / float64(weight)
		limiter = rate.NewLimiter(rate.Limit(adjustedRate), int(wrl.baseRate))
		wrl.limiters[weight] = limiter
	}
	wrl.mu.Unlock()

	return limiter.Allow()
}

// Wait waits for a weighted request
func (wrl *WeightedRateLimiter) Wait(ctx context.Context, weight int) error {
	wrl.mu.Lock()
	limiter, exists := wrl.limiters[weight]
	if !exists {
		adjustedRate := wrl.baseRate / float64(weight)
		limiter = rate.NewLimiter(rate.Limit(adjustedRate), int(wrl.baseRate))
		wrl.limiters[weight] = limiter
	}
	wrl.mu.Unlock()

	return limiter.Wait(ctx)
}
