package strategies

import (
	"fmt"
	"sync"
)

// StrategyFactory is a function that creates a new strategy instance
type StrategyFactory func() Strategy

// Registry manages strategy registration and creation
type Registry struct {
	factories map[string]StrategyFactory
	mu        sync.RWMutex
}

// NewRegistry creates a new strategy registry
func NewRegistry() *Registry {
	registry := &Registry{
		factories: make(map[string]StrategyFactory),
	}

	// Register built-in strategies
	registry.Register(StrategyTypeMomentum, func() Strategy {
		return NewMomentumStrategy()
	})

	registry.Register(StrategyTypeMarketMaking, func() Strategy {
		return NewMarketMakingStrategy()
	})

	registry.Register(StrategyTypeScalping, func() Strategy {
		return NewScalpingStrategy()
	})

	return registry
}

// Register registers a strategy factory
func (r *Registry) Register(name string, factory StrategyFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[name] = factory
}

// Create creates a new strategy instance
func (r *Registry) Create(name string) (Strategy, error) {
	r.mu.RLock()
	factory, exists := r.factories[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("strategy not found: %s", name)
	}

	return factory(), nil
}

// List returns all registered strategy names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// Has checks if a strategy is registered
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.factories[name]
	return exists
}
