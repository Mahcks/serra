package cache

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/mahcks/serra/pkg/structures"
)

// Helper function to marshal data to JSON
func marshalToJSON(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// TMDBServiceInterface defines the methods we need from the TMDB service
type TMDBServiceInterface interface {
	GetWatchProviders(mediaType string) (structures.TMDBWatchProvidersListResponse, error)
	GetWatchProviderRegions() (structures.TMDBWatchProviderRegionsResponse, error)
	GetMoviePopular(page string) (structures.TMDBMediaResponse, error)
	GetTVPopular(page string) (structures.TMDBMediaResponse, error)
	GetTrendingMedia(page string) (structures.TMDBMediaResponse, error)
}

type BackgroundCacheService struct {
	cache       *TMDBCacheService
	tmdbService TMDBServiceInterface
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewBackgroundCacheService(cacheService *TMDBCacheService, tmdbService TMDBServiceInterface) *BackgroundCacheService {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &BackgroundCacheService{
		cache:       cacheService,
		tmdbService: tmdbService,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start begins the background cache management tasks
func (b *BackgroundCacheService) Start() {
	log.Println("🗄️ Starting TMDB cache background service...")
	
	// Initial cache warming
	go b.warmStaticCache()
	
	// Periodic cleanup task (every hour)
	go b.startCleanupTask()
	
	// Periodic cache warming (every 6 hours)
	go b.startWarmingTask()
	
	log.Println("✅ TMDB cache background service started")
}

// Stop shuts down the background service
func (b *BackgroundCacheService) Stop() {
	log.Println("🛑 Stopping TMDB cache background service...")
	b.cancel()
}

// warmStaticCache preloads static data that rarely changes
func (b *BackgroundCacheService) warmStaticCache() {
	log.Println("🔥 Warming TMDB static cache...")
	
	// Warm genres (rarely change, high value)
	if _, found, _ := b.cache.GetStaticData("movie_genres"); !found {
		log.Println("📥 Caching movie genres...")
		// This would require adding a GetGenres method to your TMDB service
		// For now, we'll skip this but it's recommended to add
	}
	
	// Warm popular watch providers
	if _, found, _ := b.cache.GetStaticData("watch_providers_movie"); !found {
		log.Println("📥 Caching movie watch providers...")
		if providers, err := b.tmdbService.GetWatchProviders("movie"); err == nil {
			if jsonData, err := marshalToJSON(providers); err == nil {
				b.cache.SetStaticData("watch_providers_movie", jsonData)
				log.Println("✅ Cached movie watch providers")
			}
		}
	}
	
	if _, found, _ := b.cache.GetStaticData("watch_providers_tv"); !found {
		log.Println("📥 Caching TV watch providers...")
		if providers, err := b.tmdbService.GetWatchProviders("tv"); err == nil {
			if jsonData, err := marshalToJSON(providers); err == nil {
				b.cache.SetStaticData("watch_providers_tv", jsonData)
				log.Println("✅ Cached TV watch providers")
			}
		}
	}
	
	// Warm regions
	if _, found, _ := b.cache.GetStaticData("watch_provider_regions"); !found {
		log.Println("📥 Caching watch provider regions...")
		if regions, err := b.tmdbService.GetWatchProviderRegions(); err == nil {
			if jsonData, err := marshalToJSON(regions); err == nil {
				b.cache.SetStaticData("watch_provider_regions", jsonData)
				log.Println("✅ Cached watch provider regions")
			}
		}
	}
	
	log.Println("🔥 Static cache warming completed")
}

// startCleanupTask runs periodic cache cleanup
func (b *BackgroundCacheService) startCleanupTask() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			b.cleanupCache()
		}
	}
}

// startWarmingTask runs periodic cache warming for popular data
func (b *BackgroundCacheService) startWarmingTask() {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			b.warmPopularCache()
		}
	}
}

// cleanupCache removes expired entries and old API usage data
func (b *BackgroundCacheService) cleanupCache() {
	log.Println("🧹 Starting cache cleanup...")
	
	// Get stats before cleanup
	statsBefore, err := b.cache.GetCacheStats()
	if err == nil {
		log.Printf("📊 Cache stats before cleanup: %d total, %d valid, %d expired", 
			statsBefore["total_entries"], statsBefore["valid_entries"], statsBefore["expired_entries"])
	}
	
	// Remove expired cache entries
	if err := b.cache.CleanupExpiredCache(); err != nil {
		log.Printf("❌ Failed to cleanup expired cache: %v", err)
	}
	
	// Clean up old API usage data (older than 30 days)  
	if err := b.cache.CleanupOldAPIUsage(); err != nil {
		log.Printf("❌ Failed to cleanup old API usage: %v", err)
	}
	
	// Get stats after cleanup
	statsAfter, err := b.cache.GetCacheStats()
	if err == nil {
		log.Printf("📊 Cache stats after cleanup: %d total, %d valid", 
			statsAfter["total_entries"], statsAfter["valid_entries"])
	}
	
	log.Println("✅ Cache cleanup completed")
}

// warmPopularCache refreshes frequently accessed data
func (b *BackgroundCacheService) warmPopularCache() {
	log.Println("🔥 Warming popular cache data...")
	
	// Refresh popular movies (first page)
	if _, err := b.tmdbService.GetMoviePopular("1"); err != nil {
		log.Printf("❌ Failed to warm popular movies cache: %v", err)
	} else {
		log.Println("✅ Warmed popular movies cache")
	}
	
	// Refresh popular TV shows
	if _, err := b.tmdbService.GetTVPopular("1"); err != nil {
		log.Printf("❌ Failed to warm popular TV cache: %v", err)
	} else {
		log.Println("✅ Warmed popular TV cache")
	}
	
	// Refresh trending
	if _, err := b.tmdbService.GetTrendingMedia("1"); err != nil {
		log.Printf("❌ Failed to warm trending cache: %v", err)
	} else {
		log.Println("✅ Warmed trending cache")
	}
	
	log.Println("🔥 Popular cache warming completed")
}

// GetCacheStats returns current cache statistics
func (b *BackgroundCacheService) GetCacheStats() (map[string]interface{}, error) {
	stats, err := b.cache.GetCacheStats()
	if err != nil {
		return nil, err
	}
	
	// Add API usage info
	apiUsage, err := b.cache.GetAPIUsageToday()
	if err == nil {
		stats["api_requests_today"] = apiUsage
	}
	
	// Convert to interface map for JSON serialization
	result := make(map[string]interface{})
	for k, v := range stats {
		result[k] = v
	}
	
	return result, nil
}