package utils

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	// ErrCircuitOpen is returned when the circuit breaker is open.
	ErrCircuitOpen = errors.New("circuit breaker is open")
	// ErrTooManyRequests is returned when too many requests are made in half-open state.
	ErrTooManyRequests = errors.New("too many requests")
)

// CircuitBreakerState represents the state of a circuit breaker.
type CircuitBreakerState int

const (
	// StateClosed means the circuit is closed (normal operation).
	StateClosed CircuitBreakerState = iota
	// StateOpen means the circuit is open (blocking requests).
	StateOpen
	// StateHalfOpen means the circuit is half-open (testing recovery).
	StateHalfOpen
)

// CircuitBreakerConfig configures a circuit breaker.
type CircuitBreakerConfig struct {
	MaxFailures     int           // Max failures before opening
	Timeout         time.Duration // Duration to stay open
	HalfOpenMaxReqs int           // Max requests in half-open state
	OnStateChange   func(from, to CircuitBreakerState)
}

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	config        CircuitBreakerConfig
	state         CircuitBreakerState
	failures      int
	successes     int
	lastFailTime  time.Time
	halfOpenReqs  int
	mu            sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	if config.MaxFailures == 0 {
		config.MaxFailures = 5
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.HalfOpenMaxReqs == 0 {
		config.HalfOpenMaxReqs = 3
	}

	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

// Execute executes the given function with circuit breaker protection.
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	err := fn()

	cb.afterRequest(err)
	return err
}

// GetState returns the current state of the circuit breaker.
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	oldState := cb.state
	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.halfOpenReqs = 0

	if cb.config.OnStateChange != nil && oldState != StateClosed {
		cb.config.OnStateChange(oldState, StateClosed)
	}
}

// beforeRequest checks if a request can be made.
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil

	case StateOpen:
		if time.Since(cb.lastFailTime) > cb.config.Timeout {
			cb.setState(StateHalfOpen)
			cb.halfOpenReqs = 1
			return nil
		}
		return ErrCircuitOpen

	case StateHalfOpen:
		if cb.halfOpenReqs >= cb.config.HalfOpenMaxReqs {
			return ErrTooManyRequests
		}
		cb.halfOpenReqs++
		return nil

	default:
		return nil
	}
}

// afterRequest records the result of a request.
func (cb *CircuitBreaker) afterRequest(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
}

// onFailure handles a failed request.
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.MaxFailures {
			cb.setState(StateOpen)
		}

	case StateHalfOpen:
		cb.setState(StateOpen)
	}
}

// onSuccess handles a successful request.
func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		cb.failures = 0

	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.HalfOpenMaxReqs {
			cb.setState(StateClosed)
			cb.failures = 0
			cb.successes = 0
		}
	}
}

// setState changes the circuit breaker state and triggers callback.
func (cb *CircuitBreaker) setState(newState CircuitBreakerState) {
	oldState := cb.state
	cb.state = newState
	cb.halfOpenReqs = 0

	if cb.config.OnStateChange != nil {
		cb.config.OnStateChange(oldState, newState)
	}
}
