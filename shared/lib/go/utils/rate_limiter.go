package utils

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter.
type RateLimiter struct {
	rate       int           // Tokens per second
	burst      int           // Max tokens in bucket
	tokens     float64       // Current tokens
	lastUpdate time.Time     // Last token update time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter.
// rate is the number of tokens per second.
// burst is the maximum number of tokens that can be accumulated.
func NewRateLimiter(rate, burst int) *RateLimiter {
	if burst < rate {
		burst = rate
	}

	return &RateLimiter{
		rate:       rate,
		burst:      burst,
		tokens:     float64(burst),
		lastUpdate: time.Now(),
	}
}

// Allow checks if a request can proceed without blocking.
func (rl *RateLimiter) Allow() bool {
	return rl.AllowN(1)
}

// AllowN checks if n requests can proceed without blocking.
func (rl *RateLimiter) AllowN(n int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refillTokens()

	if rl.tokens >= float64(n) {
		rl.tokens -= float64(n)
		return true
	}

	return false
}

// Wait blocks until a request can proceed or context is canceled.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	return rl.WaitN(ctx, 1)
}

// WaitN blocks until n requests can proceed or context is canceled.
func (rl *RateLimiter) WaitN(ctx context.Context, n int) error {
	for {
		if rl.AllowN(n) {
			return nil
		}

		// Calculate wait time
		rl.mu.Lock()
		needed := float64(n) - rl.tokens
		waitTime := time.Duration(needed/float64(rl.rate)*1e9) * time.Nanosecond
		rl.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Continue to next iteration
		}
	}
}

// Reserve reserves n tokens for future use and returns a Reservation.
func (rl *RateLimiter) Reserve(n int) *Reservation {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refillTokens()

	if rl.tokens >= float64(n) {
		rl.tokens -= float64(n)
		return &Reservation{
			ok:    true,
			delay: 0,
		}
	}

	needed := float64(n) - rl.tokens
	delay := time.Duration(needed/float64(rl.rate)*1e9) * time.Nanosecond
	rl.tokens = 0

	return &Reservation{
		ok:    true,
		delay: delay,
	}
}

// refillTokens adds tokens based on elapsed time since last update.
func (rl *RateLimiter) refillTokens() {
	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate)
	rl.lastUpdate = now

	tokensToAdd := elapsed.Seconds() * float64(rl.rate)
	rl.tokens = min(rl.tokens+tokensToAdd, float64(rl.burst))
}

// Tokens returns the current number of available tokens.
func (rl *RateLimiter) Tokens() float64 {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refillTokens()
	return rl.tokens
}

// Limit returns the rate limit (tokens per second).
func (rl *RateLimiter) Limit() int {
	return rl.rate
}

// Burst returns the burst size.
func (rl *RateLimiter) Burst() int {
	return rl.burst
}

// Reservation represents a reservation for tokens.
type Reservation struct {
	ok    bool
	delay time.Duration
}

// OK returns whether the reservation is valid.
func (r *Reservation) OK() bool {
	return r.ok
}

// Delay returns the delay before the reservation becomes valid.
func (r *Reservation) Delay() time.Duration {
	return r.delay
}

// DelayFrom returns the delay from a specific time.
func (r *Reservation) DelayFrom(now time.Time) time.Duration {
	return r.delay
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
