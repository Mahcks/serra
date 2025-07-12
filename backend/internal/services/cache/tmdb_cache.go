package cache

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mahcks/serra/internal/db/repository"
)

type TMDBCacheService struct {
	db *repository.Queries
}

func NewTMDBCacheService(database *repository.Queries) *TMDBCacheService {
	return &TMDBCacheService{
		db: database,
	}
}

// Cache TTL constants
const (
	StaticDataTTL    = 30 * 24 * time.Hour // 30 days for genres, companies, etc.
	MovieDetailsTTL  = 7 * 24 * time.Hour  // 7 days for movie/show details
	SearchResultsTTL = 1 * time.Hour       // 1 hour for search results
	DynamicDataTTL   = 15 * time.Minute    // 15 minutes for trending/popular
)

// GenerateCacheKey creates a consistent cache key from endpoint and parameters
func (c *TMDBCacheService) GenerateCacheKey(endpoint string, params map[string]interface{}) string {
	// Create a consistent hash of the parameters
	paramJSON, _ := json.Marshal(params)
	hash := fmt.Sprintf("%x", md5.Sum(paramJSON))
	return fmt.Sprintf("tmdb:%s:%s", endpoint, hash)
}

// GetCachedData retrieves data from cache if it exists and is not expired
func (c *TMDBCacheService) GetCachedData(cacheKey string) ([]byte, bool, error) {
	ctx := context.Background()
	entry, err := c.db.GetCacheEntry(ctx, cacheKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil // Cache miss, not an error
		}
		return nil, false, fmt.Errorf("failed to get cache entry: %w", err)
	}

	return []byte(entry.Data), true, nil
}

// SetCachedData stores data in cache with appropriate TTL
func (c *TMDBCacheService) SetCachedData(cacheKey string, endpoint string, data []byte, ttl time.Duration) error {
	ctx := context.Background()
	expiresAt := time.Now().Add(ttl)

	err := c.db.SetCacheEntry(ctx, repository.SetCacheEntryParams{
		CacheKey:  cacheKey,
		Data:      string(data),
		Endpoint:  endpoint,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return fmt.Errorf("failed to set cache entry: %w", err)
	}

	return nil
}

// GetStaticData retrieves static TMDB data (genres, companies, etc.)
func (c *TMDBCacheService) GetStaticData(dataType string) ([]byte, bool, error) {
	ctx := context.Background()
	entry, err := c.db.GetStaticData(ctx, dataType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to get static data: %w", err)
	}

	// Check if data is older than TTL
	if entry.LastUpdated.Valid && time.Since(entry.LastUpdated.Time) > StaticDataTTL {
		return nil, false, nil // Data is stale
	}

	return []byte(entry.Data), true, nil
}

// SetStaticData stores static TMDB data
func (c *TMDBCacheService) SetStaticData(dataType string, data []byte) error {
	ctx := context.Background()
	err := c.db.SetStaticData(ctx, repository.SetStaticDataParams{
		DataType: dataType,
		Data:     string(data),
	})
	if err != nil {
		return fmt.Errorf("failed to set static data: %w", err)
	}
	return nil
}

// GetTTLForEndpoint returns appropriate TTL based on endpoint type
func (c *TMDBCacheService) GetTTLForEndpoint(endpoint string) time.Duration {
	switch {
	case endpoint == "genre/movie/list" || endpoint == "genre/tv/list":
		return StaticDataTTL
	case endpoint == "search/company" || endpoint == "watch/providers/movie" || endpoint == "watch/providers/tv":
		return StaticDataTTL
	case endpoint == "trending/all/day" || endpoint == "movie/popular" || endpoint == "tv/popular":
		return DynamicDataTTL
	case endpoint == "search/movie" || endpoint == "search/tv":
		return SearchResultsTTL
	default:
		// Default for movie/show details, recommendations, etc.
		return MovieDetailsTTL
	}
}

// TrackAPIUsage increments the API usage counter for rate limiting tracking
func (c *TMDBCacheService) TrackAPIUsage(endpoint string) error {
	ctx := context.Background()
	err := c.db.IncrementAPIUsage(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("failed to track API usage: %w", err)
	}
	return nil
}

// GetAPIUsageToday returns today's total API request count
func (c *TMDBCacheService) GetAPIUsageToday() (int64, error) {
	ctx := context.Background()
	usage, err := c.db.GetAPIUsageToday(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get API usage: %w", err)
	}
	
	// Convert interface{} to int64
	if usageInt64, ok := usage.(int64); ok {
		return usageInt64, nil
	}
	return 0, nil
}

// CleanupExpiredCache removes expired cache entries
func (c *TMDBCacheService) CleanupExpiredCache() error {
	ctx := context.Background()
	err := c.db.DeleteExpiredCache(ctx)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired cache: %w", err)
	}
	return nil
}

// GetCacheStats returns cache statistics
func (c *TMDBCacheService) GetCacheStats() (map[string]int64, error) {
	ctx := context.Background()
	stats, err := c.db.GetCacheStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache stats: %w", err)
	}

	return map[string]int64{
		"total_entries":   stats.TotalEntries,
		"valid_entries":   stats.ValidEntries,
		"expired_entries": stats.ExpiredEntries,
	}, nil
}

// InvalidateEndpoint removes all cache entries for a specific endpoint
func (c *TMDBCacheService) InvalidateEndpoint(endpoint string) error {
	ctx := context.Background()
	err := c.db.DeleteCacheByEndpoint(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("failed to invalidate endpoint cache: %w", err)
	}
	return nil
}

// CleanupOldAPIUsage removes old API usage data (older than 30 days)
func (c *TMDBCacheService) CleanupOldAPIUsage() error {
	ctx := context.Background()
	err := c.db.CleanupOldAPIUsage(ctx)
	if err != nil {
		return fmt.Errorf("failed to cleanup old API usage: %w", err)
	}
	return nil
}