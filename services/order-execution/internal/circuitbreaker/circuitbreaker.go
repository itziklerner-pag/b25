package circuitbreaker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sony/gobreaker"
)

// CircuitBreaker wraps the gobreaker implementation
type CircuitBreaker struct {
	breakers map[string]*gobreaker.CircuitBreaker
	mu       sync.RWMutex
	settings gobreaker.Settings
}

// Config holds circuit breaker configuration
type Config struct {
	MaxRequests       uint32
	Interval          time.Duration
	Timeout           time.Duration
	FailureThreshold  uint32
	SuccessThreshold  uint32
	OnStateChange     func(name string, from gobreaker.State, to gobreaker.State)
}

// DefaultConfig returns default circuit breaker configuration
func DefaultConfig() Config {
	return Config{
		MaxRequests:      3,
		Interval:         60 * time.Second,
		Timeout:          30 * time.Second,
		FailureThreshold: 5,
		SuccessThreshold: 2,
	}
}

// NewCircuitBreaker creates a new circuit breaker manager
func NewCircuitBreaker(config Config) *CircuitBreaker {
	settings := gobreaker.Settings{
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= config.FailureThreshold
		},
		OnStateChange: config.OnStateChange,
	}

	return &CircuitBreaker{
		breakers: make(map[string]*gobreaker.CircuitBreaker),
		settings: settings,
	}
}

// GetBreaker returns or creates a circuit breaker for a name
func (cb *CircuitBreaker) GetBreaker(name string) *gobreaker.CircuitBreaker {
	cb.mu.RLock()
	breaker, exists := cb.breakers[name]
	cb.mu.RUnlock()

	if exists {
		return breaker
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Double-check after acquiring write lock
	breaker, exists = cb.breakers[name]
	if exists {
		return breaker
	}

	settings := cb.settings
	settings.Name = name

	breaker = gobreaker.NewCircuitBreaker(settings)
	cb.breakers[name] = breaker

	return breaker
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(name string, fn func() (interface{}, error)) (interface{}, error) {
	breaker := cb.GetBreaker(name)
	return breaker.Execute(fn)
}

// ExecuteWithContext executes a function with context and circuit breaker protection
func (cb *CircuitBreaker) ExecuteWithContext(ctx context.Context, name string, fn func() (interface{}, error)) (interface{}, error) {
	// Check context before execution
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	breaker := cb.GetBreaker(name)

	// Check circuit breaker state
	state := breaker.State()
	if state == gobreaker.StateOpen {
		return nil, fmt.Errorf("circuit breaker %s is open", name)
	}

	// Execute with context monitoring
	type result struct {
		data interface{}
		err  error
	}

	resultChan := make(chan result, 1)

	go func() {
		data, err := breaker.Execute(fn)
		resultChan <- result{data: data, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-resultChan:
		return res.data, res.err
	}
}

// GetState returns the current state of a circuit breaker
func (cb *CircuitBreaker) GetState(name string) gobreaker.State {
	cb.mu.RLock()
	breaker, exists := cb.breakers[name]
	cb.mu.RUnlock()

	if !exists {
		return gobreaker.StateClosed
	}

	return breaker.State()
}

// Reset resets a circuit breaker
func (cb *CircuitBreaker) Reset(name string) {
	cb.mu.RLock()
	breaker, exists := cb.breakers[name]
	cb.mu.RUnlock()

	if exists {
		// Create a new breaker with the same settings
		cb.mu.Lock()
		settings := cb.settings
		settings.Name = name
		cb.breakers[name] = gobreaker.NewCircuitBreaker(settings)
		cb.mu.Unlock()
	}
}

// GetCounts returns the counts for a circuit breaker
func (cb *CircuitBreaker) GetCounts(name string) gobreaker.Counts {
	cb.mu.RLock()
	breaker, exists := cb.breakers[name]
	cb.mu.RUnlock()

	if !exists {
		return gobreaker.Counts{}
	}

	return breaker.Counts()
}

// MultiBreaker manages multiple circuit breakers for different operations
type MultiBreaker struct {
	primary   *CircuitBreaker
	secondary *CircuitBreaker
	mu        sync.RWMutex
}

// NewMultiBreaker creates a multi-level circuit breaker
func NewMultiBreaker(primaryConfig, secondaryConfig Config) *MultiBreaker {
	return &MultiBreaker{
		primary:   NewCircuitBreaker(primaryConfig),
		secondary: NewCircuitBreaker(secondaryConfig),
	}
}

// ExecutePrimary executes with primary circuit breaker
func (mb *MultiBreaker) ExecutePrimary(name string, fn func() (interface{}, error)) (interface{}, error) {
	return mb.primary.Execute(name, fn)
}

// ExecuteSecondary executes with secondary circuit breaker
func (mb *MultiBreaker) ExecuteSecondary(name string, fn func() (interface{}, error)) (interface{}, error) {
	return mb.secondary.Execute(name, fn)
}

// ExecuteWithFallback executes with primary, falls back to secondary on failure
func (mb *MultiBreaker) ExecuteWithFallback(name string, primary, fallback func() (interface{}, error)) (interface{}, error) {
	result, err := mb.primary.Execute(name+"-primary", primary)
	if err != nil {
		// Try fallback
		return mb.secondary.Execute(name+"-fallback", fallback)
	}
	return result, nil
}

// AdaptiveBreaker adjusts thresholds based on error patterns
type AdaptiveBreaker struct {
	breaker          *gobreaker.CircuitBreaker
	errorWindow      []time.Time
	windowSize       time.Duration
	adaptiveInterval time.Duration
	mu               sync.RWMutex
}

// NewAdaptiveBreaker creates an adaptive circuit breaker
func NewAdaptiveBreaker(name string, windowSize, adaptiveInterval time.Duration) *AdaptiveBreaker {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
	}

	ab := &AdaptiveBreaker{
		breaker:          gobreaker.NewCircuitBreaker(settings),
		errorWindow:      make([]time.Time, 0),
		windowSize:       windowSize,
		adaptiveInterval: adaptiveInterval,
	}

	// Start adaptive adjustment
	go ab.adapt()

	return ab
}

// Execute executes with adaptive circuit breaker
func (ab *AdaptiveBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	result, err := ab.breaker.Execute(fn)

	if err != nil {
		ab.mu.Lock()
		ab.errorWindow = append(ab.errorWindow, time.Now())
		ab.mu.Unlock()
	}

	return result, err
}

// adapt periodically adjusts thresholds based on error patterns
func (ab *AdaptiveBreaker) adapt() {
	ticker := time.NewTicker(ab.adaptiveInterval)
	defer ticker.Stop()

	for range ticker.C {
		ab.mu.Lock()

		// Remove old errors from window
		cutoff := time.Now().Add(-ab.windowSize)
		validErrors := make([]time.Time, 0)
		for _, errTime := range ab.errorWindow {
			if errTime.After(cutoff) {
				validErrors = append(validErrors, errTime)
			}
		}
		ab.errorWindow = validErrors

		// Adjust thresholds based on error rate
		// This is a simplified adaptation - production would be more sophisticated
		errorRate := float64(len(ab.errorWindow)) / ab.windowSize.Seconds()

		// Update settings based on error rate
		// (In practice, you'd need to recreate the breaker with new settings)
		_ = errorRate // Placeholder for actual adaptation logic

		ab.mu.Unlock()
	}
}

// GetErrorRate returns current error rate
func (ab *AdaptiveBreaker) GetErrorRate() float64 {
	ab.mu.RLock()
	defer ab.mu.RUnlock()

	cutoff := time.Now().Add(-ab.windowSize)
	count := 0
	for _, errTime := range ab.errorWindow {
		if errTime.After(cutoff) {
			count++
		}
	}

	return float64(count) / ab.windowSize.Seconds()
}
