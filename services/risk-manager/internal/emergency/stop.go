package emergency

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// StopStatus represents the current emergency stop state
type StopStatus struct {
	IsStopped      bool
	StoppedAt      time.Time
	StopReason     string
	TriggeredBy    string
	OrdersCanceled int
	PositionsClosed int
	Completed      bool
	CompletedAt    time.Time
	mu             sync.RWMutex
}

// StopManager manages emergency stop functionality
type StopManager struct {
	status         *StopStatus
	logger         *zap.Logger
	alertPublisher AlertPublisher
	mu             sync.RWMutex
}

// AlertPublisher defines the interface for publishing alerts
type AlertPublisher interface {
	PublishEmergencyAlert(ctx context.Context, reason, triggeredBy string) error
}

// NewStopManager creates a new emergency stop manager
func NewStopManager(logger *zap.Logger, alertPublisher AlertPublisher) *StopManager {
	return &StopManager{
		status: &StopStatus{
			IsStopped: false,
		},
		logger:         logger,
		alertPublisher: alertPublisher,
	}
}

// Trigger initiates an emergency stop
func (m *StopManager) Trigger(ctx context.Context, reason, triggeredBy string, force bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already stopped
	if m.status.IsStopped && !force {
		return fmt.Errorf("emergency stop already active")
	}

	m.logger.Error("EMERGENCY STOP TRIGGERED",
		zap.String("reason", reason),
		zap.String("triggered_by", triggeredBy),
		zap.Bool("force", force),
	)

	// Update status
	m.status.IsStopped = true
	m.status.StoppedAt = time.Now()
	m.status.StopReason = reason
	m.status.TriggeredBy = triggeredBy
	m.status.OrdersCanceled = 0
	m.status.PositionsClosed = 0
	m.status.Completed = false

	// Publish emergency alert
	if err := m.alertPublisher.PublishEmergencyAlert(ctx, reason, triggeredBy); err != nil {
		m.logger.Error("failed to publish emergency alert", zap.Error(err))
		// Continue with emergency stop even if alert fails
	}

	return nil
}

// IsActive returns whether emergency stop is currently active
func (m *StopManager) IsActive() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status.IsStopped
}

// GetStatus returns the current emergency stop status
func (m *StopManager) GetStatus() StopStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent race conditions
	return StopStatus{
		IsStopped:       m.status.IsStopped,
		StoppedAt:       m.status.StoppedAt,
		StopReason:      m.status.StopReason,
		TriggeredBy:     m.status.TriggeredBy,
		OrdersCanceled:  m.status.OrdersCanceled,
		PositionsClosed: m.status.PositionsClosed,
		Completed:       m.status.Completed,
		CompletedAt:     m.status.CompletedAt,
	}
}

// UpdateProgress updates the emergency stop progress
func (m *StopManager) UpdateProgress(ordersCanceled, positionsClosed int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.status.OrdersCanceled = ordersCanceled
	m.status.PositionsClosed = positionsClosed

	m.logger.Info("emergency stop progress",
		zap.Int("orders_canceled", ordersCanceled),
		zap.Int("positions_closed", positionsClosed),
	)
}

// Complete marks the emergency stop as completed
func (m *StopManager) Complete() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.status.Completed = true
	m.status.CompletedAt = time.Now()

	duration := m.status.CompletedAt.Sub(m.status.StoppedAt)

	m.logger.Info("emergency stop completed",
		zap.Duration("duration", duration),
		zap.Int("orders_canceled", m.status.OrdersCanceled),
		zap.Int("positions_closed", m.status.PositionsClosed),
	)
}

// ReEnable re-enables trading after emergency stop
func (m *StopManager) ReEnable(authorizedBy, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.status.IsStopped {
		return fmt.Errorf("emergency stop is not active")
	}

	if !m.status.Completed {
		return fmt.Errorf("emergency stop has not completed yet")
	}

	m.logger.Info("re-enabling trading after emergency stop",
		zap.String("authorized_by", authorizedBy),
		zap.String("reason", reason),
		zap.Duration("downtime", time.Since(m.status.StoppedAt)),
	)

	// Reset status
	m.status.IsStopped = false
	m.status.StopReason = ""
	m.status.TriggeredBy = ""
	m.status.Completed = false

	return nil
}

// ShouldBlockOrders returns whether orders should be blocked
func (m *StopManager) ShouldBlockOrders() bool {
	return m.IsActive()
}

// CircuitBreaker implements circuit breaker pattern for risk violations
type CircuitBreaker struct {
	threshold      int
	window         time.Duration
	violations     []time.Time
	mu             sync.RWMutex
	onTrip         func()
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(threshold int, window time.Duration, onTrip func()) *CircuitBreaker {
	return &CircuitBreaker{
		threshold:  threshold,
		window:     window,
		violations: make([]time.Time, 0),
		onTrip:     onTrip,
	}
}

// RecordViolation records a violation and checks if breaker should trip
func (cb *CircuitBreaker) RecordViolation() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	// Remove old violations outside the window
	validViolations := make([]time.Time, 0)
	for _, t := range cb.violations {
		if now.Sub(t) <= cb.window {
			validViolations = append(validViolations, t)
		}
	}

	// Add current violation
	validViolations = append(validViolations, now)
	cb.violations = validViolations

	// Check if threshold exceeded
	if len(cb.violations) >= cb.threshold {
		if cb.onTrip != nil {
			go cb.onTrip()
		}
		return true
	}

	return false
}

// Reset resets the circuit breaker
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.violations = make([]time.Time, 0)
}

// Count returns the current violation count
func (cb *CircuitBreaker) Count() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	now := time.Now()
	count := 0
	for _, t := range cb.violations {
		if now.Sub(t) <= cb.window {
			count++
		}
	}
	return count
}
