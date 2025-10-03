package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/b25/services/risk-manager/internal/limits"
	"github.com/redis/go-redis/v9"
)

const (
	policyCacheKey = "risk:policies"
	priceCacheKey  = "market:prices:%s" // market:prices:{symbol}
)

// PolicyCache manages policy caching
type PolicyCache struct {
	redis      *redis.Client
	ttl        time.Duration
	localCache *localPolicyCache
	mu         sync.RWMutex
}

// localPolicyCache provides in-memory fallback
type localPolicyCache struct {
	policies  []*limits.Policy
	expiresAt time.Time
	mu        sync.RWMutex
}

// NewPolicyCache creates a new policy cache
func NewPolicyCache(redis *redis.Client, ttl time.Duration) *PolicyCache {
	return &PolicyCache{
		redis: redis,
		ttl:   ttl,
		localCache: &localPolicyCache{
			policies: make([]*limits.Policy, 0),
		},
	}
}

// SetPolicies stores policies in cache
func (c *PolicyCache) SetPolicies(ctx context.Context, policies []*limits.Policy) error {
	// Update local cache
	c.localCache.mu.Lock()
	c.localCache.policies = policies
	c.localCache.expiresAt = time.Now().Add(c.ttl)
	c.localCache.mu.Unlock()

	// Update Redis cache
	data, err := json.Marshal(policies)
	if err != nil {
		return fmt.Errorf("marshal policies: %w", err)
	}

	if err := c.redis.Set(ctx, policyCacheKey, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("set redis cache: %w", err)
	}

	return nil
}

// GetPolicies retrieves policies from cache
func (c *PolicyCache) GetPolicies(ctx context.Context) ([]*limits.Policy, error) {
	// Try local cache first
	c.localCache.mu.RLock()
	if time.Now().Before(c.localCache.expiresAt) && len(c.localCache.policies) > 0 {
		policies := c.localCache.policies
		c.localCache.mu.RUnlock()
		return policies, nil
	}
	c.localCache.mu.RUnlock()

	// Try Redis cache
	data, err := c.redis.Get(ctx, policyCacheKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("cache miss")
		}
		return nil, fmt.Errorf("get redis cache: %w", err)
	}

	var policies []*limits.Policy
	if err := json.Unmarshal(data, &policies); err != nil {
		return nil, fmt.Errorf("unmarshal policies: %w", err)
	}

	// Update local cache
	c.localCache.mu.Lock()
	c.localCache.policies = policies
	c.localCache.expiresAt = time.Now().Add(c.ttl)
	c.localCache.mu.Unlock()

	return policies, nil
}

// Invalidate clears the policy cache
func (c *PolicyCache) Invalidate(ctx context.Context) error {
	c.localCache.mu.Lock()
	c.localCache.policies = nil
	c.localCache.expiresAt = time.Time{}
	c.localCache.mu.Unlock()

	return c.redis.Del(ctx, policyCacheKey).Err()
}

// MarketPriceCache manages market price caching
type MarketPriceCache struct {
	redis *redis.Client
	ttl   time.Duration
}

// NewMarketPriceCache creates a new market price cache
func NewMarketPriceCache(redis *redis.Client, ttl time.Duration) *MarketPriceCache {
	return &MarketPriceCache{
		redis: redis,
		ttl:   ttl,
	}
}

// GetPrice retrieves price for a symbol
func (c *MarketPriceCache) GetPrice(ctx context.Context, symbol string) (float64, error) {
	key := fmt.Sprintf(priceCacheKey, symbol)
	price, err := c.redis.Get(ctx, key).Float64()
	if err != nil {
		if err == redis.Nil {
			return 0, fmt.Errorf("price not found for symbol %s", symbol)
		}
		return 0, fmt.Errorf("get price from redis: %w", err)
	}
	return price, nil
}

// GetPrices retrieves prices for multiple symbols
func (c *MarketPriceCache) GetPrices(ctx context.Context, symbols []string) (map[string]float64, error) {
	if len(symbols) == 0 {
		return make(map[string]float64), nil
	}

	// Build keys
	keys := make([]string, len(symbols))
	for i, symbol := range symbols {
		keys[i] = fmt.Sprintf(priceCacheKey, symbol)
	}

	// Batch get
	values, err := c.redis.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("mget prices from redis: %w", err)
	}

	prices := make(map[string]float64)
	for i, val := range values {
		if val == nil {
			continue
		}

		// Parse price
		var price float64
		switch v := val.(type) {
		case string:
			if _, err := fmt.Sscanf(v, "%f", &price); err != nil {
				continue
			}
		case float64:
			price = v
		default:
			continue
		}

		prices[symbols[i]] = price
	}

	return prices, nil
}

// SetPrice stores a price in cache
func (c *MarketPriceCache) SetPrice(ctx context.Context, symbol string, price float64) error {
	key := fmt.Sprintf(priceCacheKey, symbol)
	return c.redis.Set(ctx, key, price, c.ttl).Err()
}

// AccountStateCache manages account state caching
type AccountStateCache struct {
	redis *redis.Client
	ttl   time.Duration
}

// NewAccountStateCache creates a new account state cache
func NewAccountStateCache(redis *redis.Client, ttl time.Duration) *AccountStateCache {
	return &AccountStateCache{
		redis: redis,
		ttl:   ttl,
	}
}

// GetAccountState retrieves cached account state
func (c *AccountStateCache) GetAccountState(ctx context.Context, accountID string) (map[string]interface{}, error) {
	key := fmt.Sprintf("account:state:%s", accountID)
	data, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("cache miss")
		}
		return nil, fmt.Errorf("get account state: %w", err)
	}

	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshal account state: %w", err)
	}

	return state, nil
}

// SetAccountState stores account state in cache
func (c *AccountStateCache) SetAccountState(ctx context.Context, accountID string, state map[string]interface{}) error {
	key := fmt.Sprintf("account:state:%s", accountID)
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal account state: %w", err)
	}

	return c.redis.Set(ctx, key, data, c.ttl).Err()
}
