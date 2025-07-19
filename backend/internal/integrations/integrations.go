package integrations

import (
	"log/slog"

	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/internal/integrations/cached"
	"github.com/mahcks/serra/internal/integrations/emby"
	"github.com/mahcks/serra/internal/integrations/jellystat"
	"github.com/mahcks/serra/internal/integrations/radarr"
	"github.com/mahcks/serra/internal/integrations/rottentomatoes"
	"github.com/mahcks/serra/internal/integrations/sonarr"
	"github.com/mahcks/serra/internal/integrations/tmdb"
	"github.com/mahcks/serra/internal/services/cache"
)

type Integration struct {
	Radarr          radarr.Service
	Sonarr          sonarr.Service
	Jellystat       jellystat.Service
	Emby            emby.Service
	TMDB            tmdb.Service
	CacheService    *cache.TMDBCacheService
	BackgroundCache *cache.BackgroundCacheService
	RottenTomatoes  rottentomatoes.Service
}

func New(gctx global.Context) *Integration {
	var tmdbService tmdb.Service
	var cacheService *cache.TMDBCacheService
	var backgroundCache *cache.BackgroundCacheService

	tmdbAPIKey := gctx.Crate().Config.Get().TMDB.APIKey.String()

	if tmdbAPIKey != "" {
		// Initialize base TMDB service
		baseService, err := tmdb.New(tmdb.Options{
			BaseURL: "https://api.themoviedb.org/3",
			APIKey:  tmdbAPIKey,
		})
		if err != nil {
			slog.Warn("Failed to initialize TMDB service", "error", err)
			tmdbService = nil
		} else {
			// Initialize cache service
			cacheService = cache.NewTMDBCacheService(gctx.Crate().Sqlite.Query())

			// Wrap base service with caching
			tmdbService = cached.NewTMDBService(baseService, cacheService)

			// Initialize background cache service
			backgroundCache = cache.NewBackgroundCacheService(cacheService, tmdbService)
			backgroundCache.Start()

			slog.Info("TMDB service with caching initialized successfully")
		}
	} else {
		slog.Info("TMDB API key not configured, skipping TMDB service initialization")
		tmdbService = nil
	}

	return &Integration{
		Radarr:          radarr.New(gctx.Crate().Sqlite.Query()),
		Sonarr:          sonarr.New(gctx.Crate().Sqlite.Query()),
		Jellystat:       jellystat.New(gctx),
		Emby:            emby.New(gctx),
		TMDB:            tmdbService,
		CacheService:    cacheService,
		BackgroundCache: backgroundCache,
		RottenTomatoes:  rottentomatoes.NewService(),
	}
}

// Shutdown gracefully stops all background services
func (i *Integration) Shutdown() {
	if i.BackgroundCache != nil {
		i.BackgroundCache.Stop()
	}
}
