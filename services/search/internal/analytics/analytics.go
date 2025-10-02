package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/yourorg/b25/services/search/internal/config"
	"github.com/yourorg/b25/services/search/pkg/models"
)

// Analytics tracks search analytics
type Analytics struct {
	redis  *redis.Client
	config *config.AnalyticsConfig
	ttl    time.Duration
	logger *zap.Logger
}

// NewAnalytics creates a new analytics tracker
func NewAnalytics(redisClient *redis.Client, cfg *config.AnalyticsConfig, ttl time.Duration, logger *zap.Logger) *Analytics {
	return &Analytics{
		redis:  redisClient,
		config: cfg,
		ttl:    ttl,
		logger: logger,
	}
}

// TrackSearch records a search query
func (a *Analytics) TrackSearch(ctx context.Context, analytics *models.SearchAnalytics) error {
	if !a.config.Enabled || !a.config.TrackSearches {
		return nil
	}

	// Store individual search record
	key := fmt.Sprintf("search:analytics:%s:%d", analytics.Query, analytics.Timestamp.Unix())
	data, err := json.Marshal(analytics)
	if err != nil {
		return fmt.Errorf("failed to marshal analytics: %w", err)
	}

	if err := a.redis.Set(ctx, key, data, a.ttl).Err(); err != nil {
		return fmt.Errorf("failed to store analytics: %w", err)
	}

	// Increment search count for popular queries
	popularKey := "search:popular:" + analytics.Query
	pipe := a.redis.Pipeline()
	pipe.Incr(ctx, popularKey)
	pipe.Expire(ctx, popularKey, a.ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		a.logger.Warn("Failed to update popular search", zap.Error(err))
	}

	// Track by index
	if analytics.Index != "" {
		indexKey := fmt.Sprintf("search:index:%s:count", analytics.Index)
		pipe := a.redis.Pipeline()
		pipe.Incr(ctx, indexKey)
		pipe.Expire(ctx, indexKey, a.ttl)
		if _, err := pipe.Exec(ctx); err != nil {
			a.logger.Warn("Failed to update index search count", zap.Error(err))
		}
	}

	// Track latency metrics
	latencyKey := fmt.Sprintf("search:latency:%s", time.Now().Format("2006-01-02"))
	pipe = a.redis.Pipeline()
	pipe.RPush(ctx, latencyKey, analytics.Latency)
	pipe.Expire(ctx, latencyKey, 24*time.Hour)
	// Keep only last 1000 latencies
	pipe.LTrim(ctx, latencyKey, -1000, -1)
	if _, err := pipe.Exec(ctx); err != nil {
		a.logger.Warn("Failed to update latency metrics", zap.Error(err))
	}

	return nil
}

// TrackClick records a click on a search result
func (a *Analytics) TrackClick(ctx context.Context, click *models.ClickAnalytics) error {
	if !a.config.Enabled || !a.config.TrackClicks {
		return nil
	}

	// Store click record
	key := fmt.Sprintf("search:click:%s:%s:%d", click.Query, click.DocumentID, click.Timestamp.Unix())
	data, err := json.Marshal(click)
	if err != nil {
		return fmt.Errorf("failed to marshal click: %w", err)
	}

	if err := a.redis.Set(ctx, key, data, a.ttl).Err(); err != nil {
		return fmt.Errorf("failed to store click: %w", err)
	}

	// Increment click-through rate counter
	ctrKey := fmt.Sprintf("search:ctr:%s:%s", click.Query, click.DocumentID)
	pipe := a.redis.Pipeline()
	pipe.Incr(ctx, ctrKey)
	pipe.Expire(ctx, ctrKey, a.ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		a.logger.Warn("Failed to update CTR", zap.Error(err))
	}

	return nil
}

// GetPopularSearches returns the most popular search queries
func (a *Analytics) GetPopularSearches(ctx context.Context, limit int) ([]models.PopularSearch, error) {
	if !a.config.Enabled || !a.config.TrackSearches {
		return nil, fmt.Errorf("analytics not enabled")
	}

	// Scan for popular search keys
	pattern := "search:popular:*"
	keys, err := a.scanKeys(ctx, pattern, limit*2) // Get more keys to sort and limit
	if err != nil {
		return nil, fmt.Errorf("failed to scan keys: %w", err)
	}

	if len(keys) == 0 {
		return []models.PopularSearch{}, nil
	}

	// Get counts for all keys
	pipe := a.redis.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get counts: %w", err)
	}

	// Parse results
	searches := make([]models.PopularSearch, 0, len(keys))
	for i, cmd := range cmds {
		count, err := cmd.Int64()
		if err != nil {
			continue
		}

		// Extract query from key: "search:popular:{query}"
		query := keys[i][16:] // Remove "search:popular:" prefix

		searches = append(searches, models.PopularSearch{
			Query:       query,
			SearchCount: count,
			LastUsed:    time.Now(), // Approximate
		})
	}

	// Sort by count (descending) and limit
	for i := 0; i < len(searches); i++ {
		for j := i + 1; j < len(searches); j++ {
			if searches[j].SearchCount > searches[i].SearchCount {
				searches[i], searches[j] = searches[j], searches[i]
			}
		}
	}

	if len(searches) > limit {
		searches = searches[:limit]
	}

	return searches, nil
}

// GetSearchStats returns search statistics
func (a *Analytics) GetSearchStats(ctx context.Context) (map[string]interface{}, error) {
	if !a.config.Enabled {
		return nil, fmt.Errorf("analytics not enabled")
	}

	stats := make(map[string]interface{})

	// Get total searches by index
	indexPattern := "search:index:*:count"
	indexKeys, err := a.scanKeys(ctx, indexPattern, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to scan index keys: %w", err)
	}

	indexStats := make(map[string]int64)
	for _, key := range indexKeys {
		count, err := a.redis.Get(ctx, key).Int64()
		if err != nil {
			continue
		}
		// Extract index name: "search:index:{index}:count"
		indexName := key[13 : len(key)-6] // Remove prefix and suffix
		indexStats[indexName] = count
	}
	stats["searches_by_index"] = indexStats

	// Get average latency
	latencyKey := fmt.Sprintf("search:latency:%s", time.Now().Format("2006-01-02"))
	latencies, err := a.redis.LRange(ctx, latencyKey, 0, -1).Result()
	if err == nil && len(latencies) > 0 {
		var total int64
		for _, l := range latencies {
			var latency int64
			if err := json.Unmarshal([]byte(l), &latency); err == nil {
				total += latency
			}
		}
		stats["avg_latency_ms"] = float64(total) / float64(len(latencies))
		stats["total_searches_today"] = len(latencies)
	}

	// Get popular searches
	popular, err := a.GetPopularSearches(ctx, 10)
	if err == nil {
		stats["popular_searches"] = popular
	}

	return stats, nil
}

// GetClickThroughRate returns CTR for a query
func (a *Analytics) GetClickThroughRate(ctx context.Context, query string) (float64, error) {
	if !a.config.Enabled || !a.config.TrackClicks {
		return 0, fmt.Errorf("analytics not enabled")
	}

	// Get search count
	searchKey := "search:popular:" + query
	searchCount, err := a.redis.Get(ctx, searchKey).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get search count: %w", err)
	}

	// Get click count
	ctrPattern := fmt.Sprintf("search:ctr:%s:*", query)
	ctrKeys, err := a.scanKeys(ctx, ctrPattern, 1000)
	if err != nil {
		return 0, fmt.Errorf("failed to scan CTR keys: %w", err)
	}

	var totalClicks int64
	for _, key := range ctrKeys {
		count, err := a.redis.Get(ctx, key).Int64()
		if err == nil {
			totalClicks += count
		}
	}

	if searchCount == 0 {
		return 0, nil
	}

	return float64(totalClicks) / float64(searchCount), nil
}

// CleanupOldData removes analytics data older than retention period
func (a *Analytics) CleanupOldData(ctx context.Context) error {
	if !a.config.Enabled {
		return nil
	}

	cutoff := time.Now().AddDate(0, 0, -a.config.RetentionDays)

	// Cleanup search analytics
	pattern := "search:analytics:*"
	keys, err := a.scanKeys(ctx, pattern, 10000)
	if err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	deleted := 0
	for _, key := range keys {
		data, err := a.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var analytics models.SearchAnalytics
		if err := json.Unmarshal([]byte(data), &analytics); err != nil {
			continue
		}

		if analytics.Timestamp.Before(cutoff) {
			if err := a.redis.Del(ctx, key).Err(); err == nil {
				deleted++
			}
		}
	}

	// Cleanup click analytics
	pattern = "search:click:*"
	keys, err = a.scanKeys(ctx, pattern, 10000)
	if err != nil {
		return fmt.Errorf("failed to scan click keys: %w", err)
	}

	for _, key := range keys {
		data, err := a.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var click models.ClickAnalytics
		if err := json.Unmarshal([]byte(data), &click); err != nil {
			continue
		}

		if click.Timestamp.Before(cutoff) {
			if err := a.redis.Del(ctx, key).Err(); err == nil {
				deleted++
			}
		}
	}

	a.logger.Info("Cleaned up old analytics data",
		zap.Int("deleted", deleted),
		zap.Time("cutoff", cutoff),
	)

	return nil
}

// scanKeys scans Redis keys matching a pattern
func (a *Analytics) scanKeys(ctx context.Context, pattern string, limit int) ([]string, error) {
	var keys []string
	var cursor uint64

	for {
		var scanKeys []string
		var err error

		scanKeys, cursor, err = a.redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 || len(keys) >= limit {
			break
		}
	}

	if len(keys) > limit {
		keys = keys[:limit]
	}

	return keys, nil
}

// StartCleanupWorker starts a background worker to cleanup old data
func (a *Analytics) StartCleanupWorker(ctx context.Context) {
	if !a.config.Enabled {
		return
	}

	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := a.CleanupOldData(ctx); err != nil {
					a.logger.Error("Failed to cleanup old analytics data", zap.Error(err))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	a.logger.Info("Analytics cleanup worker started")
}
