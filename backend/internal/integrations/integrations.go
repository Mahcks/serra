package integrations

import (
	"log/slog"

	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/internal/integrations/emby"
	"github.com/mahcks/serra/internal/integrations/jellystat"
	"github.com/mahcks/serra/internal/integrations/radarr"
	"github.com/mahcks/serra/internal/integrations/sonarr"
	"github.com/mahcks/serra/internal/integrations/tmdb"
)

type Integration struct {
	Radarr    radarr.Service
	Sonarr    sonarr.Service
	Jellystat jellystat.Service
	Emby      emby.Service
	TMDB      tmdb.Service
}

func New(gctx global.Context) *Integration {
	var tmdbService tmdb.Service
	tmdbAPIKey := gctx.Crate().Config.Get().TMDB.APIKey.String()

	if tmdbAPIKey != "" {
		service, err := tmdb.New(tmdb.Options{
			BaseURL: "https://api.themoviedb.org/3",
			APIKey:  tmdbAPIKey,
		})
		if err != nil {
			slog.Warn("Failed to initialize TMDB service", "error", err)
			tmdbService = nil
		} else {
			tmdbService = service
			slog.Info("TMDB service initialized successfully")
		}
	} else {
		slog.Info("TMDB API key not configured, skipping TMDB service initialization")
		tmdbService = nil
	}

	return &Integration{
		Radarr:    radarr.New(gctx.Crate().Sqlite.Query()),
		Sonarr:    sonarr.New(gctx.Crate().Sqlite.Query()),
		Jellystat: jellystat.New(gctx),
		Emby:      emby.New(gctx),
		TMDB:      tmdbService,
	}
}
