package breaker

import (
	"fmt"
	"sync"

	"github.com/b25/api-gateway/internal/config"
	"github.com/b25/api-gateway/pkg/logger"
	"github.com/b25/api-gateway/pkg/metrics"
	"github.com/sony/gobreaker"
)

// Manager manages circuit breakers for different services
type Manager struct {
	breakers map[string]*gobreaker.CircuitBreaker
	config   config.CircuitBreakerConfig
	log      *logger.Logger
	metrics  *metrics.Collector
	mu       sync.RWMutex
}

// NewManager creates a new circuit breaker manager
func NewManager(cfg config.CircuitBreakerConfig, log *logger.Logger, m *metrics.Collector) *Manager {
	return &Manager{
		breakers: make(map[string]*gobreaker.CircuitBreaker),
		config:   cfg,
		log:      log,
		metrics:  m,
	}
}

// GetBreaker returns a circuit breaker for a service, creating it if necessary
func (m *Manager) GetBreaker(service string) *gobreaker.CircuitBreaker {
	if !m.config.Enabled {
		return nil
	}

	m.mu.RLock()
	breaker, exists := m.breakers[service]
	m.mu.RUnlock()

	if exists {
		return breaker
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	breaker, exists = m.breakers[service]
	if exists {
		return breaker
	}

	// Get service-specific config or use defaults
	breakerConfig := m.getServiceConfig(service)

	// Create new circuit breaker
	settings := gobreaker.Settings{
		Name:        service,
		MaxRequests: breakerConfig.MaxRequests,
		Interval:    breakerConfig.Interval,
		Timeout:     breakerConfig.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			m.log.Warnw("Circuit breaker state changed",
				"service", name,
				"from", from.String(),
				"to", to.String(),
			)

			// Update metrics based on state
			var stateValue float64
			switch to {
			case gobreaker.StateClosed:
				stateValue = 0
			case gobreaker.StateHalfOpen:
				stateValue = 1
			case gobreaker.StateOpen:
				stateValue = 2
			}
			m.metrics.SetCircuitBreakerState(name, stateValue)
		},
	}

	breaker = gobreaker.NewCircuitBreaker(settings)
	m.breakers[service] = breaker

	return breaker
}

// Execute executes a function through a circuit breaker
func (m *Manager) Execute(service string, fn func() (interface{}, error)) (interface{}, error) {
	if !m.config.Enabled {
		return fn()
	}

	breaker := m.GetBreaker(service)
	if breaker == nil {
		return fn()
	}

	result, err := breaker.Execute(fn)
	if err != nil {
		if err == gobreaker.ErrOpenState {
			m.log.Warnw("Circuit breaker open",
				"service", service,
			)
			return nil, fmt.Errorf("service %s is currently unavailable", service)
		}
		return nil, err
	}

	return result, nil
}

// GetState returns the state of a circuit breaker
func (m *Manager) GetState(service string) gobreaker.State {
	m.mu.RLock()
	defer m.mu.RUnlock()

	breaker, exists := m.breakers[service]
	if !exists {
		return gobreaker.StateClosed
	}

	return breaker.State()
}

// GetCounts returns the counts for a circuit breaker
func (m *Manager) GetCounts(service string) gobreaker.Counts {
	m.mu.RLock()
	defer m.mu.RUnlock()

	breaker, exists := m.breakers[service]
	if !exists {
		return gobreaker.Counts{}
	}

	return breaker.Counts()
}

// getServiceConfig gets configuration for a specific service
func (m *Manager) getServiceConfig(service string) config.ServiceBreakerConfig {
	if svcConfig, exists := m.config.Services[service]; exists {
		return svcConfig
	}

	// Return default config
	return config.ServiceBreakerConfig{
		MaxRequests: m.config.MaxRequests,
		Interval:    m.config.Interval,
		Timeout:     m.config.Timeout,
	}
}
