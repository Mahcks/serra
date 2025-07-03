package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/internal/integrations"
	"github.com/mahcks/serra/internal/websocket"
	"github.com/mahcks/serra/pkg/downloadclient"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

// Pattern arrays for content detection
var (
	tvPatterns = []string{
		"s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8", "s9", "s10", "s11", "s12", "s13", "s14", "s15", "s16", "s17", "s18", "s19", "s20",
		"e0", "e1", "e2", "e3", "e4", "e5", "e6", "e7", "e8", "e9", "e10", "e11", "e12", "e13", "e14", "e15", "e16", "e17", "e18", "e19", "e20",
		"season", "episode", "s01e", "s02e", "s03e", "s04e", "s05e", "s06e", "s07e", "s08e", "s09e", "s10e", "s11e", "s12e", "s13e", "s14e", "s15e", "s16e", "s17e", "s18e", "s19e", "s20e",
	}
	moviePatterns = []string{
		"1080p", "720p", "4k", "2160p", "bluray", "web-dl", "hdtv", "dvdrip", "brrip",
		"x264", "x265", "h264", "h265", "aac", "ac3", "dts",
	}

	// Shared HTTP client with proper timeouts
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}
)

// DownloadPoller is a refactored version that uses the new client architecture
type DownloadPoller struct {
	gctx          global.Context
	ticker        *time.Ticker
	stopChan      chan struct{}
	clientManager *integrations.DownloadClientManager

	// Metrics
	lastPollTime   time.Time
	pollCount      int64
	errorCount     int64
	downloadsFound int64
	cleanupCount   int64

	// Caching
	radarrCache map[string]struct {
		Movies map[int]struct {
			TmdbID int    `json:"tmdbId"`
			Title  string `json:"title"`
		}
		LastUpdated time.Time
	}
	sonarrCache map[string]struct {
		Series map[int]struct {
			TmdbID int    `json:"tmdbId"`
			Title  string `json:"title"`
		}
		Episodes    map[int]sonarrEpisode
		LastUpdated time.Time
	}
	cacheMutex sync.RWMutex

	// Adaptive polling
	lastDownloadCount int
	pollInterval      time.Duration
	baseInterval      time.Duration

	// Circuit breakers
	circuitBreakers map[string]*circuitBreaker
	cbMutex         sync.RWMutex

	lastCleanupTime  time.Time // For improved cleanup scheduling
	lastCacheCleanup time.Time // For cache cleanup scheduling
	started          bool
}

// Download represents a download item for internal use
type Download struct {
	ID           string
	Title        string
	TorrentTitle string
	Source       string
	TmdbID       *int64
	Progress     float64
	TimeLeft     *string
	Status       *string
	Hash         *string
}

// sonarrEpisode represents episode data from Sonarr
type sonarrEpisode struct {
	ID            int    `json:"id"`
	SeriesID      int    `json:"seriesId"`
	SeasonNumber  int    `json:"seasonNumber"`
	EpisodeNumber int    `json:"episodeNumber"`
	Title         string `json:"title"`
}

// circuitBreaker constants
const (
	cbClosed   = "closed"
	cbOpen     = "open"
	cbHalfOpen = "half-open"
)

// circuitBreaker implements a simple circuit breaker pattern
type circuitBreaker struct {
	failureCount    int64
	lastFailureTime int64 // Unix timestamp
	state           int32 // 0=closed, 1=open, 2=half-open
	mutex           sync.RWMutex
}

// Circuit breaker states
const (
	cbClosedState   int32 = 0
	cbOpenState     int32 = 1
	cbHalfOpenState int32 = 2
)

// NewDownloadPoller creates a new DownloadPoller instance
func NewDownloadPoller(gctx global.Context) (*DownloadPoller, error) {
	dp := &DownloadPoller{
		gctx:          gctx,
		ticker:        time.NewTicker(15 * time.Second),
		stopChan:      make(chan struct{}),
		clientManager: integrations.NewDownloadClientManager(),
		pollInterval:  15 * time.Second,
		baseInterval:  15 * time.Second,
		radarrCache: make(map[string]struct {
			Movies map[int]struct {
				TmdbID int    `json:"tmdbId"`
				Title  string `json:"title"`
			}
			LastUpdated time.Time
		}),
		sonarrCache: make(map[string]struct {
			Series map[int]struct {
				TmdbID int    `json:"tmdbId"`
				Title  string `json:"title"`
			}
			Episodes    map[int]sonarrEpisode
			LastUpdated time.Time
		}),
		circuitBreakers:  make(map[string]*circuitBreaker),
		lastCleanupTime:  time.Now(),
		lastCacheCleanup: time.Now(),
		started:          false,
	}

	// Initialize download clients
	if err := dp.initializeClients(); err != nil {
		return nil, err
	}

	return dp, nil
}

// Name returns the job name
func (dp *DownloadPoller) Name() structures.Job {
	return structures.JobDownloadPoller
}

func (dp *DownloadPoller) Trigger() error {
	go dp.pollCombined()
	return nil
}

// Start begins the download poller loop (non-blocking)
func (dp *DownloadPoller) Start() {
	if dp.started {
		return
	}
	dp.started = true
	dp.pollCombined()
	go dp.startPolling()
}

func (dp *DownloadPoller) initializeClients() error {
	// Query download clients from database
	clients, err := dp.gctx.Crate().Sqlite.Query().GetDownloadClients(context.Background())
	if err != nil {
		slog.Error("Failed to get download clients from database", "error", err)
		return err
	}

	slog.Info("Found download clients in database", "count", len(clients))
	for _, client := range clients {
		slog.Info("Download client", "id", client.ID, "type", client.Type, "name", client.Name, "host", client.Host, "port", client.Port)
	}

	// Initialize all clients
	err = dp.clientManager.InitializeClients(clients)
	if err != nil {
		slog.Error("Failed to initialize download clients", "error", err)
		return err
	}

	slog.Info("Successfully initialized download clients", "count", len(clients))
	return nil
}

func (dp *DownloadPoller) startPolling() {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("Poller goroutine panicked, restarting", "error", r)
				}
			}()
			select {
			case <-dp.ticker.C:
				dp.pollCombined()
			case <-dp.stopChan:
				dp.ticker.Stop()
				return
			}
		}()
	}
}

func (dp *DownloadPoller) Stop(ctx context.Context) error {
	close(dp.stopChan)
	return dp.clientManager.CloseAll(ctx)
}

func (dp *DownloadPoller) pollCombined() {
	ctx := context.Background()

	slog.Debug("====================START OF DOWNLOAD JOB====================")

	// 1. Fetch all downloads from all clients (for matching)
	slog.Debug("STEP 1: Fetching downloads from all clients")
	allClientDownloads, err := dp.clientManager.GetAllDownloads(ctx)
	if err != nil {
		slog.Error("Failed to get downloads from clients", "error", err)
		dp.errorCount++
		slog.Debug("====================END OF DOWNLOAD JOB (ERROR)====================")
		return
	}
	slog.Debug("STEP 1 COMPLETE: Fetched downloads from all clients", "count", len(allClientDownloads))

	// Build a map for fast lookup by hash and by name
	slog.Debug("STEP 2: Building download lookup maps")
	downloadsByHash := make(map[string]downloadclient.Item)
	downloadsByName := make(map[string]downloadclient.Item)
	for _, d := range allClientDownloads {
		if d.Hash != "" {
			downloadsByHash[strings.ToLower(d.Hash)] = d
		}
		if d.Name != "" {
			downloadsByName[strings.ToLower(d.Name)] = d
		}
	}
	slog.Debug("STEP 2 COMPLETE: Built lookup maps", "byHash", len(downloadsByHash), "byName", len(downloadsByName))

	var allEnrichedDownloads []Download

	// 2. For each Radarr instance, fetch queue and enrich
	slog.Debug("STEP 3: Processing Radarr instances")
	radarrInstances, err := dp.gctx.Crate().Sqlite.Query().GetArrServiceByType(ctx, "radarr")
	if err != nil {
		slog.Error("Failed to fetch Radarr instances", "error", err)
	} else {
		slog.Debug("Found Radarr instances", "count", len(radarrInstances))
		for i, radarr := range radarrInstances {
			slog.Debug("Processing Radarr instance", "index", i+1, "total", len(radarrInstances), "name", radarr.Name)
			queue, err := fetchRadarrQueue(ctx, radarr.BaseUrl, radarr.ApiKey)
			if err != nil {
				slog.Error("Failed to fetch Radarr queue", "name", radarr.Name, "error", err)
				continue
			}
			slog.Debug("Fetched Radarr queue", "name", radarr.Name, "queueSize", len(queue))

			for _, item := range queue {
				// Try to match to a download client item
				var matched *downloadclient.Item
				if item.DownloadID != "" {
					if m, ok := downloadsByHash[strings.ToLower(item.DownloadID)]; ok {
						matched = &m
					}
				}
				if matched == nil && item.Title != "" {
					if m, ok := downloadsByName[strings.ToLower(item.Title)]; ok {
						matched = &m
					}
				}

				// Fetch movie details
				movie, err := fetchRadarrMovie(ctx, radarr.BaseUrl, radarr.ApiKey, item.MovieID)
				if err != nil {
					slog.Info("Failed to fetch Radarr movie details", "movieID", item.MovieID, "error", err)
					continue
				}

				progress := 0.0
				timeLeft := item.TimeLeft
				status := item.Status
				hash := ""
				if matched != nil {
					progress = matched.Progress
					if matched.TimeLeft != "" {
						timeLeft = matched.TimeLeft
					}
					if matched.Status != "" {
						status = matched.Status
					}
					hash = matched.Hash
				}

				uniqueID := fmt.Sprintf("%s_%s", radarr.ID, item.DownloadID)
				allEnrichedDownloads = append(allEnrichedDownloads, Download{
					ID:           uniqueID,
					Title:        movie.Title,
					TorrentTitle: item.Title,
					Source:       "radarr",
					TmdbID:       utils.PtrInt64(int64(movie.TmdbID)),
					Progress:     progress,
					TimeLeft:     utils.PtrString(timeLeft),
					Status:       utils.PtrString(status),
					Hash:         utils.PtrString(hash),
				})
			}
		}
	}
	slog.Debug("STEP 3 COMPLETE: Processed Radarr instances", "enrichedDownloads", len(allEnrichedDownloads))

	// 3. For each Sonarr instance, fetch queue and enrich
	slog.Debug("STEP 4: Processing Sonarr instances")
	sonarrInstances, err := dp.gctx.Crate().Sqlite.Query().GetArrServiceByType(ctx, "sonarr")
	if err != nil {
		slog.Error("Failed to fetch Sonarr instances", "error", err)
	} else {
		slog.Debug("Found Sonarr instances", "count", len(sonarrInstances))
		for i, sonarr := range sonarrInstances {
			slog.Debug("Processing Sonarr instance", "index", i+1, "total", len(sonarrInstances), "name", sonarr.Name)
			queue, err := fetchSonarrQueue(ctx, sonarr.BaseUrl, sonarr.ApiKey)
			if err != nil {
				slog.Error("Failed to fetch Sonarr queue", "name", sonarr.Name, "error", err)
				continue
			}
			slog.Debug("Fetched Sonarr queue", "name", sonarr.Name, "queueSize", len(queue))

			for _, item := range queue {
				// Try to match to a download client item
				var matched *downloadclient.Item
				if item.DownloadID != "" {
					if m, ok := downloadsByHash[strings.ToLower(item.DownloadID)]; ok {
						matched = &m
					}
				}
				if matched == nil && item.Title != "" {
					if m, ok := downloadsByName[strings.ToLower(item.Title)]; ok {
						matched = &m
					}
				}

				// Fetch series/episode details
				series, err := fetchSonarrSeries(ctx, sonarr.BaseUrl, sonarr.ApiKey, item.SeriesID)
				if err != nil {
					slog.Info("Failed to fetch Sonarr series details", "seriesID", item.SeriesID, "error", err)
					continue
				}

				episode, err := fetchSonarrEpisode(ctx, sonarr.BaseUrl, sonarr.ApiKey, item.EpisodeID)
				if err != nil {
					slog.Info("Failed to fetch Sonarr episode details", "episodeID", item.EpisodeID, "error", err)
					continue
				}

				progress := 0.0
				timeLeft := item.TimeLeft
				status := item.Status
				hash := ""
				if matched != nil {
					progress = matched.Progress
					if matched.TimeLeft != "" {
						timeLeft = matched.TimeLeft
					}
					if matched.Status != "" {
						status = matched.Status
					}
					hash = matched.Hash
				}

				uniqueID := fmt.Sprintf("%s_%s", sonarr.ID, item.DownloadID)
				var name string
				if episode.SeasonNumber == 0 && episode.EpisodeNumber == 0 {
					name = fmt.Sprintf("%s Season Pack", series.Title)
				} else {
					name = fmt.Sprintf("%s S%dxE%d - %s", series.Title, episode.SeasonNumber, episode.EpisodeNumber, episode.Title)
				}
				allEnrichedDownloads = append(allEnrichedDownloads, Download{
					ID:           uniqueID,
					Title:        name,
					TorrentTitle: item.Title,
					Source:       "sonarr",
					TmdbID:       utils.PtrInt64(int64(series.TmdbID)),
					Progress:     progress,
					TimeLeft:     utils.PtrString(timeLeft),
					Status:       utils.PtrString(status),
					Hash:         utils.PtrString(hash),
				})
			}
		}
	}
	slog.Debug("STEP 4 COMPLETE: Processed Sonarr instances", "enrichedDownloads", len(allEnrichedDownloads))

	// 4. Store downloads in database
	slog.Debug("STEP 5: Storing downloads in database")
	dp.storeDownloads(allEnrichedDownloads)
	slog.Debug("STEP 5 COMPLETE: Stored downloads in database")

	// 5. Broadcast via WebSocket (only active downloads)
	slog.Debug("STEP 6: Broadcasting via WebSocket")

	// Filter out completed downloads for WebSocket broadcast
	var activeDownloads []Download
	for _, d := range allEnrichedDownloads {
		status := utils.DerefString(d.Status)
		// Only broadcast downloads that are not completed
		if status != "completed" {
			activeDownloads = append(activeDownloads, d)
		}
	}

	if len(activeDownloads) > 0 {
		var batch []structures.DownloadProgressPayload
		for _, d := range activeDownloads {
			batch = append(batch, structures.DownloadProgressPayload{
				ID:           d.ID,
				Title:        d.Title,
				TorrentTitle: d.TorrentTitle,
				Source:       d.Source,
				TMDBID:       d.TmdbID,
				TvDBID:       nil, // Not used in current implementation
				Hash:         utils.DerefString(d.Hash),
				Progress:     d.Progress,
				TimeLeft:     utils.DerefString(d.TimeLeft),
				Status:       utils.DerefString(d.Status),
				LastUpdated:  time.Now().Format(time.RFC3339),
			})
		}
		websocket.BroadcastToAll(structures.OpcodeDownloadProgressBatch, structures.DownloadProgressBatchPayload{
			Downloads: batch,
		})
		slog.Info("Found active downloads", "count", len(activeDownloads), "total", len(allEnrichedDownloads))
		slog.Debug("STEP 6 COMPLETE: Broadcasted via WebSocket", "batchSize", len(batch), "filteredOut", len(allEnrichedDownloads)-len(activeDownloads))
	} else {
		slog.Debug("STEP 6 COMPLETE: No active downloads to broadcast")
	}

	// Clean up completed downloads from database
	slog.Debug("STEP 7: Cleaning up completed downloads")
	dp.cleanupCompletedDownloads(allEnrichedDownloads)
	slog.Debug("STEP 7 COMPLETE: Cleaned up completed downloads")

	// Clean up old missing downloads every 25 minutes
	if time.Since(dp.lastCleanupTime) > 25*time.Minute {
		slog.Debug("STEP 8: Cleaning up old missing downloads")
		dp.cleanupOldMissingDownloads()
		dp.cleanupCount++
		dp.lastCleanupTime = time.Now()
		slog.Debug("STEP 8 COMPLETE: Cleaned up old missing downloads")
	} else {
		slog.Debug("STEP 8: Skipping old missing downloads cleanup (not time yet)")
	}

	// Clean up cache every 30 minutes
	if time.Since(dp.lastCacheCleanup) > 30*time.Minute {
		slog.Debug("STEP 8.5: Cleaning up cache")
		dp.cleanupCache()
		dp.lastCacheCleanup = time.Now()
		slog.Debug("STEP 8.5 COMPLETE: Cleaned up cache")
	}

	// Update metrics
	slog.Debug("STEP 9: Updating metrics")
	dp.lastPollTime = time.Now()
	dp.pollCount++
	dp.downloadsFound += int64(len(allEnrichedDownloads))

	// Adaptive polling: adjust interval based on activity
	dp.adjustPollInterval(len(allEnrichedDownloads))
	slog.Debug("STEP 9 COMPLETE: Updated metrics", "pollCount", dp.pollCount, "downloadsFound", dp.downloadsFound, "pollInterval", dp.pollInterval)

	// Log metrics every 10th poll
	if dp.pollCount%10 == 0 {
		slog.Info(" download poller metrics",
			"polls", dp.pollCount,
			"errors", dp.errorCount,
			"downloads_found", dp.downloadsFound,
			"cleanups", dp.cleanupCount,
			"poll_interval", dp.pollInterval,
			"last_poll", dp.lastPollTime.Format(time.RFC3339))
	}

	slog.Debug("====================END OF DOWNLOAD JOB====================")
}

// enrichDownloads enriches download items with metadata from Radarr/Sonarr
func (dp *DownloadPoller) enrichDownloads(downloads []downloadclient.Item) []Download {
	var enrichedDownloads []Download

	for _, item := range downloads {
		// Try to match with Radarr/Sonarr data
		enriched := dp.enrichDownloadItem(item)
		if enriched != nil {
			enrichedDownloads = append(enrichedDownloads, *enriched)
		}
	}

	return enrichedDownloads
}

// enrichDownloadItem attempts to enrich a download item with metadata from Radarr/Sonarr
func (dp *DownloadPoller) enrichDownloadItem(item downloadclient.Item) *Download {
	// Try multiple matching strategies in order of reliability
	matchedDownload := dp.matchWithRadarr(item)
	if matchedDownload != nil {
		return matchedDownload
	}

	matchedDownload = dp.matchWithSonarr(item)
	if matchedDownload != nil {
		return matchedDownload
	}

	// Determine source based on client ID or other indicators
	source := dp.determineSource(item)

	// If no match found, create a basic download entry
	download := &Download{
		ID:           item.ID,
		Title:        item.Name,
		TorrentTitle: item.Name,
		Source:       source,
		Progress:     item.Progress,
		TimeLeft:     utils.PtrString(item.TimeLeft),
		Status:       utils.PtrString(item.Status),
		Hash:         utils.PtrString(item.Hash),
	}

	return download
}

// determineSource attempts to determine the source of a download based on various indicators
func (dp *DownloadPoller) determineSource(item downloadclient.Item) string {
	// Check by category first (most reliable indicator)
	if item.Category != "" {
		lowerCategory := strings.ToLower(item.Category)
		if strings.Contains(lowerCategory, "radarr") || strings.Contains(lowerCategory, "movies") {
			return "radarr"
		}
		if strings.Contains(lowerCategory, "sonarr") || strings.Contains(lowerCategory, "tv") || strings.Contains(lowerCategory, "series") {
			return "sonarr"
		}
	}

	// Check by title patterns
	if item.Name == "" {
		return "manual"
	}
	lowerTitle := strings.ToLower(item.Name)

	// Look for TV show patterns (season/episode indicators)
	if utils.MatchesAnyPattern(lowerTitle, tvPatterns, false) {
		return "sonarr"
	}

	// If it has movie patterns but no TV patterns, likely a movie
	if utils.MatchesAnyPattern(lowerTitle, moviePatterns, false) {
		return "radarr"
	}

	// Default fallback - try to make an educated guess based on file extensions
	if strings.Contains(lowerTitle, ".mkv") || strings.Contains(lowerTitle, ".mp4") || strings.Contains(lowerTitle, ".avi") {
		// If it has video extensions but no clear TV/movie indicators, default to manual
		return "manual"
	}

	return "manual"
}

// matchWithRadarr attempts to match a download with Radarr movie data
func (dp *DownloadPoller) matchWithRadarr(item downloadclient.Item) *Download {
	cb := dp.getCircuitBreaker(fmt.Sprintf("radarr_cache_%s", item.ID))
	if !cb.canExecute() {
		return nil
	}
	dp.cacheMutex.RLock()
	cached, exists := dp.radarrCache[item.ID]
	dp.cacheMutex.RUnlock()
	if exists && time.Since(cached.LastUpdated) < 5*time.Minute {
		for id, movie := range cached.Movies {
			if dp.isMovieMatch(item, movie) {
				uniqueID := fmt.Sprintf("%s_%s", item.ID, id)
				return &Download{
					ID:           uniqueID,
					Title:        movie.Title,
					TorrentTitle: item.Name,
					Source:       "radarr",
					TmdbID:       utils.PtrInt64(int64(movie.TmdbID)),
					Progress:     item.Progress,
					TimeLeft:     utils.PtrString(item.TimeLeft),
					Status:       utils.PtrString(item.Status),
					Hash:         utils.PtrString(item.Hash),
				}
			}
		}
	}

	if dp.isLikelyMovie(item) {
		uniqueID := fmt.Sprintf("radarr_%s", item.ID)
		return &Download{
			ID:           uniqueID,
			Title:        dp.extractMovieTitle(item.Name),
			TorrentTitle: item.Name,
			Source:       "radarr",
			Progress:     item.Progress,
			TimeLeft:     utils.PtrString(item.TimeLeft),
			Status:       utils.PtrString(item.Status),
			Hash:         utils.PtrString(item.Hash),
		}
	}

	return nil
}

// matchWithSonarr attempts to match a download with Sonarr series data
func (dp *DownloadPoller) matchWithSonarr(item downloadclient.Item) *Download {
	cb := dp.getCircuitBreaker(fmt.Sprintf("sonarr_cache_%s", item.ID))
	if !cb.canExecute() {
		return nil
	}
	dp.cacheMutex.RLock()
	cached, exists := dp.sonarrCache[item.ID]
	dp.cacheMutex.RUnlock()
	if exists && time.Since(cached.LastUpdated) < 5*time.Minute {
		for id, series := range cached.Series {
			if dp.isSeriesMatch(item, series) {
				uniqueID := fmt.Sprintf("%s_%s", item.ID, id)
				return &Download{
					ID:           uniqueID,
					Title:        series.Title,
					TorrentTitle: item.Name,
					Source:       "sonarr",
					TmdbID:       utils.PtrInt64(int64(series.TmdbID)),
					Progress:     item.Progress,
					TimeLeft:     utils.PtrString(item.TimeLeft),
					Status:       utils.PtrString(item.Status),
					Hash:         utils.PtrString(item.Hash),
				}
			}
		}
		for id, episode := range cached.Episodes {
			if dp.isEpisodeMatch(item, episode) {
				series, exists := cached.Series[episode.SeriesID]
				if exists {
					uniqueID := fmt.Sprintf("%s_%s", item.ID, id)
					return &Download{
						ID:           uniqueID,
						Title:        fmt.Sprintf("%s S%02dE%02d - %s", series.Title, episode.SeasonNumber, episode.EpisodeNumber, episode.Title),
						TorrentTitle: item.Name,
						Source:       "sonarr",
						TmdbID:       utils.PtrInt64(int64(series.TmdbID)),
						Progress:     item.Progress,
						TimeLeft:     utils.PtrString(item.TimeLeft),
						Status:       utils.PtrString(item.Status),
						Hash:         utils.PtrString(item.Hash),
					}
				}
			}
		}
	}

	if dp.isLikelyTVShow(item) {
		uniqueID := fmt.Sprintf("sonarr_%s", item.ID)
		return &Download{
			ID:           uniqueID,
			Title:        dp.extractTVShowTitle(item.Name),
			TorrentTitle: item.Name,
			Source:       "sonarr",
			Progress:     item.Progress,
			TimeLeft:     utils.PtrString(item.TimeLeft),
			Status:       utils.PtrString(item.Status),
			Hash:         utils.PtrString(item.Hash),
		}
	}

	return nil
}

// isMovieMatch determines if a download item matches a Radarr movie
func (dp *DownloadPoller) isMovieMatch(item downloadclient.Item, movie struct {
	TmdbID int    `json:"tmdbId"`
	Title  string `json:"title"`
}) bool {
	// Strategy 1: Hash matching (most reliable)
	if item.Hash != "" {
		// Check if this hash is associated with this movie in Radarr
		// This would require querying Radarr's queue/history API
		// For now, we'll skip hash matching as it requires additional API calls
	}

	// Strategy 2: Title similarity matching
	downloadTitle := strings.ToLower(item.Name)
	movieTitle := strings.ToLower(movie.Title)

	// Remove common file extensions and quality indicators
	downloadTitle = dp.cleanTitle(downloadTitle)

	// Calculate similarity score
	similarity := dp.calculateTitleSimilarity(downloadTitle, movieTitle)

	// If similarity is high enough, consider it a match
	if similarity >= 0.8 {
		slog.Debug("Movie match found by title similarity",
			"download", item.Name,
			"movie", movie.Title,
			"similarity", similarity)
		return true
	}

	// Strategy 3: Year matching
	year := dp.extractYear(downloadTitle)
	if year > 0 {
		// You could enhance this by fetching movie year from TMDb
		// For now, we'll skip year matching
	}

	return false
}

// isSeriesMatch determines if a download item matches a Sonarr series
func (dp *DownloadPoller) isSeriesMatch(item downloadclient.Item, series struct {
	TmdbID int    `json:"tmdbId"`
	Title  string `json:"title"`
}) bool {
	downloadTitle := strings.ToLower(item.Name)
	seriesTitle := strings.ToLower(series.Title)

	// Remove common file extensions and quality indicators
	downloadTitle = dp.cleanTitle(downloadTitle)

	// Calculate similarity score
	similarity := dp.calculateTitleSimilarity(downloadTitle, seriesTitle)

	// For series, we need a higher threshold as titles might be similar
	if similarity >= 0.85 {
		slog.Debug("Series match found by title similarity",
			"download", item.Name,
			"series", series.Title,
			"similarity", similarity)
		return true
	}

	return false
}

// isEpisodeMatch determines if a download item matches a specific Sonarr episode
func (dp *DownloadPoller) isEpisodeMatch(item downloadclient.Item, episode sonarrEpisode) bool {
	// Look for season/episode patterns like S01E02, 1x02, etc.
	seasonPattern := fmt.Sprintf("s%02d", episode.SeasonNumber)
	episodePattern := fmt.Sprintf("e%02d", episode.EpisodeNumber)

	if utils.ContainsIgnoreCase(item.Name, seasonPattern) && utils.ContainsIgnoreCase(item.Name, episodePattern) {
		slog.Debug("Episode match found by season/episode pattern",
			"download", item.Name,
			"episode", fmt.Sprintf("S%02dE%02d", episode.SeasonNumber, episode.EpisodeNumber))
		return true
	}

	// Alternative patterns like 1x02, 01x02
	altSeasonPattern := fmt.Sprintf("%dx", episode.SeasonNumber)
	altEpisodePattern := fmt.Sprintf("x%02d", episode.EpisodeNumber)

	if utils.ContainsIgnoreCase(item.Name, altSeasonPattern) && utils.ContainsIgnoreCase(item.Name, altEpisodePattern) {
		slog.Debug("Episode match found by alternative season/episode pattern",
			"download", item.Name,
			"episode", fmt.Sprintf("%dx%02d", episode.SeasonNumber, episode.EpisodeNumber))
		return true
	}

	return false
}

// cleanTitle removes common file extensions and quality indicators
func (dp *DownloadPoller) cleanTitle(title string) string {
	// Remove file extensions
	extensions := []string{".mkv", ".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm"}
	if utils.MatchesAnyPattern(title, extensions, true) {
		for _, ext := range extensions {
			title = strings.ReplaceAll(title, ext, "")
		}
	}

	// Remove quality indicators
	qualities := []string{"1080p", "720p", "480p", "2160p", "4k", "hdtv", "web-dl", "bluray", "dvdrip", "hddvd", "brrip"}
	if utils.MatchesAnyPattern(title, qualities, true) {
		for _, quality := range qualities {
			title = strings.ReplaceAll(title, quality, "")
		}
	}

	// Remove release groups (usually in brackets or parentheses)
	title = regexp.MustCompile(`\[.*?\]`).ReplaceAllString(title, "")
	title = regexp.MustCompile(`\(.*?\)`).ReplaceAllString(title, "")

	// Remove extra whitespace and normalize
	title = strings.TrimSpace(title)
	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")

	return title
}

// calculateTitleSimilarity calculates how similar two titles are (0.0 to 1.0)
func (dp *DownloadPoller) calculateTitleSimilarity(title1, title2 string) float64 {
	// Simple word-based similarity
	words1 := strings.Fields(title1)
	words2 := strings.Fields(title2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	// Count common words
	common := 0
	for _, word1 := range words1 {
		for _, word2 := range words2 {
			if word1 == word2 {
				common++
				break
			}
		}
	}

	// Calculate Jaccard similarity
	total := len(words1) + len(words2) - common
	if total == 0 {
		return 0.0
	}

	return float64(common) / float64(total)
}

// extractYear attempts to extract a year from a title
func (dp *DownloadPoller) extractYear(title string) int {
	// Look for 4-digit years (1900-2099)
	yearRegex := regexp.MustCompile(`(19|20)\d{2}`)
	matches := yearRegex.FindString(title)
	if matches != "" {
		if year, err := strconv.Atoi(matches); err == nil {
			return year
		}
	}
	return 0
}

// storeDownloads stores downloads in the database
func (dp *DownloadPoller) storeDownloads(downloads []Download) {
	slog.Debug("Storing downloads in database", "count", len(downloads))

	for _, download := range downloads {
		// Handle nullable fields properly
		var tmdbID sql.NullInt64
		if download.TmdbID != nil {
			tmdbID = utils.NewNullInt64(*download.TmdbID, true)
		}

		var hash sql.NullString
		if download.Hash != nil {
			hash = utils.NewNullString(*download.Hash)
		}

		var timeLeft sql.NullString
		if download.TimeLeft != nil {
			timeLeft = utils.NewNullString(*download.TimeLeft)
		}

		var status sql.NullString
		if download.Status != nil {
			status = utils.NewNullString(*download.Status)
		}

		err := dp.gctx.Crate().Sqlite.Query().UpsertDownloadQueue(context.Background(), repository.UpsertDownloadQueueParams{
			ID:           download.ID,
			Title:        download.Title,
			TorrentTitle: download.TorrentTitle,
			Source:       download.Source,
			TmdbID:       tmdbID,
			Hash:         hash,
			Progress:     utils.NewNullFloat64(download.Progress, true),
			TimeLeft:     timeLeft,
			Status:       status,
		})
		if err != nil {
			slog.Error("Failed to upsert download", "id", download.ID, "error", err)
		}
	}
}

// adjustPollInterval adjusts the polling interval based on download activity
func (dp *DownloadPoller) adjustPollInterval(currentDownloadCount int) {
	// If downloads increased, poll more frequently
	if currentDownloadCount > dp.lastDownloadCount {
		newInterval := dp.pollInterval - 5*time.Second
		if newInterval < 5*time.Second {
			newInterval = 5 * time.Second // Minimum 5 seconds
		}
		dp.pollInterval = newInterval
	} else if currentDownloadCount == 0 && dp.lastDownloadCount > 0 {
		// If downloads went from some to none, gradually increase interval
		newInterval := dp.pollInterval + 10*time.Second
		if newInterval > 60*time.Second {
			newInterval = 60 * time.Second // Maximum 60 seconds
		}
		dp.pollInterval = newInterval
	} else if currentDownloadCount == 0 && dp.lastDownloadCount == 0 {
		// If no downloads for a while, use longer interval
		if dp.pollInterval < 30*time.Second {
			dp.pollInterval = 30 * time.Second
		}
	}

	// Update ticker if interval changed
	if dp.ticker != nil {
		dp.ticker.Reset(dp.pollInterval)
	}

	dp.lastDownloadCount = currentDownloadCount
}

// cleanupCompletedDownloads removes downloads that are no longer active from the database
func (dp *DownloadPoller) cleanupCompletedDownloads(activeDownloads []Download) {
	// Get all downloads from database
	downloads, err := dp.gctx.Crate().Sqlite.Query().ListDownloads(context.Background())
	if err != nil {
		slog.Error("Failed to fetch downloads for cleanup", "error", err)
		return
	}

	// Create a set of active download IDs
	activeIDs := make(map[string]bool)
	for _, download := range activeDownloads {
		activeIDs[download.ID] = true
	}

	// Remove downloads that are no longer active
	for _, dbDownload := range downloads {
		if !activeIDs[dbDownload.ID] {
			err := dp.gctx.Crate().Sqlite.Query().DeleteDownload(context.Background(), dbDownload.ID)
			if err != nil {
				slog.Error("Failed to delete completed download", "id", dbDownload.ID, "error", err)
			} else {
				slog.Debug("Cleaned up completed download", "id", dbDownload.ID, "title", dbDownload.Title)
			}
		}
	}
}

// cleanupOldMissingDownloads removes downloads that have been missing for too long
func (dp *DownloadPoller) cleanupOldMissingDownloads() {
	// Get downloads that have been missing for more than 24 hours
	downloads, err := dp.gctx.Crate().Sqlite.Query().GetOldMissingDownloads(context.Background())
	if err != nil {
		slog.Error("Failed to fetch old missing downloads for cleanup", "error", err)
		return
	}

	for _, download := range downloads {
		err := dp.gctx.Crate().Sqlite.Query().DeleteDownload(context.Background(), download.ID)
		if err != nil {
			slog.Error("Failed to delete old missing download", "id", download.ID, "error", err)
		} else {
			lastUpdated := utils.NullableTime{NullTime: download.LastUpdated}
			age := time.Since(lastUpdated.Or(time.Time{}))
			slog.Info("Cleaned up old missing download", "id", download.ID, "title", download.Title, "age", age)
		}
	}

	if len(downloads) > 0 {
		slog.Info("Cleaned up old missing downloads", "count", len(downloads))
	}
}

// cleanupCache removes old cache entries to prevent memory leaks
func (dp *DownloadPoller) cleanupCache() {
	dp.cacheMutex.Lock()
	defer dp.cacheMutex.Unlock()

	const maxCacheAge = 2 * time.Hour
	now := time.Now()

	// Clean up Radarr cache
	for instanceID, cache := range dp.radarrCache {
		if now.Sub(cache.LastUpdated) > maxCacheAge {
			delete(dp.radarrCache, instanceID)
			slog.Debug("Cleaned up stale Radarr cache", "instanceID", instanceID, "age", now.Sub(cache.LastUpdated))
		}
	}

	// Clean up Sonarr cache
	for instanceID, cache := range dp.sonarrCache {
		if now.Sub(cache.LastUpdated) > maxCacheAge {
			delete(dp.sonarrCache, instanceID)
			slog.Debug("Cleaned up stale Sonarr cache", "instanceID", instanceID, "age", now.Sub(cache.LastUpdated))
		}
	}

	slog.Debug("Cache cleanup completed", "radarr_entries", len(dp.radarrCache), "sonarr_entries", len(dp.sonarrCache))
}

// getCachedRadarrData gets cached movie data or fetches fresh data if cache is stale
func (dp *DownloadPoller) getCachedRadarrData(instanceID, baseURL, apiKey string) (map[int]struct {
	TmdbID int    `json:"tmdbId"`
	Title  string `json:"title"`
}, error) {
	cb := dp.getCircuitBreaker(fmt.Sprintf("radarr_cache_%s", instanceID))
	if !cb.canExecute() {
		return nil, fmt.Errorf("circuit breaker open for Radarr cache")
	}
	dp.cacheMutex.RLock()
	cached, exists := dp.radarrCache[instanceID]
	dp.cacheMutex.RUnlock()
	if exists && time.Since(cached.LastUpdated) < 5*time.Minute {
		return cached.Movies, nil
	}
	moviesRaw, err := dp.fetchRadarrMovies(context.Background(), baseURL, apiKey)
	if err != nil {
		cb.recordFailure()
		return nil, err
	}
	// Convert moviesRaw to map[int]struct{TmdbID int; Title string}
	movies := make(map[int]struct {
		TmdbID int    `json:"tmdbId"`
		Title  string `json:"title"`
	})
	for id, m := range moviesRaw {
		movies[id] = struct {
			TmdbID int    `json:"tmdbId"`
			Title  string `json:"title"`
		}{TmdbID: m.TmdbID, Title: m.Title}
	}
	dp.cacheMutex.Lock()
	dp.radarrCache[instanceID] = struct {
		Movies map[int]struct {
			TmdbID int    `json:"tmdbId"`
			Title  string `json:"title"`
		}
		LastUpdated time.Time
	}{
		Movies:      movies,
		LastUpdated: time.Now(),
	}
	dp.cacheMutex.Unlock()
	cb.recordSuccess()
	return movies, nil
}

// fetchRadarrMovies fetches all movies from a Radarr instance
func (dp *DownloadPoller) fetchRadarrMovies(ctx context.Context, baseURL, apiKey string) (map[int]struct {
	TmdbID      int    `json:"tmdbId"`
	Title       string `json:"title"`
	LastUpdated time.Time
}, error) {
	url := utils.BuildURL(baseURL, "/api/v3/movie", map[string]string{"apikey": apiKey})
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if !utils.IsHTTPSuccess(resp.StatusCode) {
		return nil, fmt.Errorf("Radarr API request failed: %s", resp.Status)
	}
	var movies []struct {
		ID     int    `json:"id"`
		Title  string `json:"title"`
		TmdbID int    `json:"tmdbId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&movies); err != nil {
		return nil, err
	}
	result := make(map[int]struct {
		TmdbID      int    `json:"tmdbId"`
		Title       string `json:"title"`
		LastUpdated time.Time
	})
	now := time.Now()
	for _, movie := range movies {
		result[movie.ID] = struct {
			TmdbID      int    `json:"tmdbId"`
			Title       string `json:"title"`
			LastUpdated time.Time
		}{
			TmdbID:      movie.TmdbID,
			Title:       movie.Title,
			LastUpdated: now,
		}
	}
	slog.Debug("Fetched movies from Radarr", "count", len(movies), "url", baseURL)
	return result, nil
}

// getCachedSonarrData gets cached series/episode data or fetches fresh data if cache is stale
func (dp *DownloadPoller) getCachedSonarrData(instanceID, baseURL, apiKey string) (map[int]struct {
	TmdbID int    `json:"tmdbId"`
	Title  string `json:"title"`
}, map[int]sonarrEpisode, error) {
	cb := dp.getCircuitBreaker(fmt.Sprintf("sonarr_cache_%s", instanceID))
	if !cb.canExecute() {
		return nil, nil, fmt.Errorf("circuit breaker open for Sonarr cache")
	}
	dp.cacheMutex.RLock()
	cached, exists := dp.sonarrCache[instanceID]
	dp.cacheMutex.RUnlock()
	if exists && time.Since(cached.LastUpdated) < 5*time.Minute {
		return cached.Series, cached.Episodes, nil
	}
	series, episodes, err := dp.fetchSonarrData(context.Background(), baseURL, apiKey)
	if err != nil {
		cb.recordFailure()
		return nil, nil, err
	}
	dp.cacheMutex.Lock()
	dp.sonarrCache[instanceID] = struct {
		Series map[int]struct {
			TmdbID int    `json:"tmdbId"`
			Title  string `json:"title"`
		}
		Episodes    map[int]sonarrEpisode
		LastUpdated time.Time
	}{
		Series:      series,
		Episodes:    episodes,
		LastUpdated: time.Now(),
	}
	dp.cacheMutex.Unlock()
	cb.recordSuccess()
	slog.Debug("Fetched fresh Sonarr data", "instance", instanceID, "series_count", len(series), "episodes_count", len(episodes))
	return series, episodes, nil
}

// fetchSonarrData fetches all series and episodes from a Sonarr instance
func (dp *DownloadPoller) fetchSonarrData(ctx context.Context, baseURL, apiKey string) (map[int]struct {
	TmdbID int    `json:"tmdbId"`
	Title  string `json:"title"`
}, map[int]sonarrEpisode, error) {
	seriesURL := utils.BuildURL(baseURL, "/api/v3/series", map[string]string{"apikey": apiKey})
	seriesReq, err := http.NewRequestWithContext(ctx, "GET", seriesURL, nil)
	if err != nil {
		return nil, nil, err
	}
	seriesResp, err := httpClient.Do(seriesReq)
	if err != nil {
		return nil, nil, err
	}
	defer seriesResp.Body.Close()
	if !utils.IsHTTPSuccess(seriesResp.StatusCode) {
		return nil, nil, fmt.Errorf("Sonarr series API request failed: %s", seriesResp.Status)
	}
	var seriesList []struct {
		ID     int    `json:"id"`
		Title  string `json:"title"`
		TmdbID int    `json:"tmdbId"`
	}
	if err := json.NewDecoder(seriesResp.Body).Decode(&seriesList); err != nil {
		return nil, nil, err
	}
	seriesMap := make(map[int]struct {
		TmdbID int    `json:"tmdbId"`
		Title  string `json:"title"`
	})
	for _, s := range seriesList {
		seriesMap[s.ID] = struct {
			TmdbID int    `json:"tmdbId"`
			Title  string `json:"title"`
		}{
			TmdbID: s.TmdbID,
			Title:  s.Title,
		}
	}
	// TODO: Batch episode fetches if Sonarr API supports it (current logic fetches all episodes for all series, which may be slow for large libraries)
	episodesMap := make(map[int]sonarrEpisode)
	for _, s := range seriesList {
		episodesURL := utils.BuildURL(baseURL, "/api/v3/episode", map[string]string{
			"seriesId": fmt.Sprintf("%d", s.ID),
			"apikey":   apiKey,
		})
		episodesReq, err := http.NewRequestWithContext(ctx, "GET", episodesURL, nil)
		if err != nil {
			continue // Skip this series if we can't create request
		}
		episodesResp, err := httpClient.Do(episodesReq)
		if err != nil {
			continue // Skip this series if request fails
		}
		if !utils.IsHTTPSuccess(episodesResp.StatusCode) {
			episodesResp.Body.Close()
			continue // Skip this series if request fails
		}
		var episodes []sonarrEpisode
		if err := json.NewDecoder(episodesResp.Body).Decode(&episodes); err != nil {
			episodesResp.Body.Close()
			continue // Skip this series if parsing fails
		}
		episodesResp.Body.Close()
		for _, episode := range episodes {
			episodesMap[episode.ID] = episode
		}
	}
	slog.Debug("Fetched data from Sonarr", "series_count", len(seriesList), "episodes_count", len(episodesMap), "url", baseURL)
	return seriesMap, episodesMap, nil
}

// getCircuitBreaker gets or creates a circuit breaker for a service
func (dp *DownloadPoller) getCircuitBreaker(serviceID string) *circuitBreaker {
	dp.cbMutex.Lock()
	defer dp.cbMutex.Unlock()

	if dp.circuitBreakers == nil {
		dp.circuitBreakers = make(map[string]*circuitBreaker)
	}

	if cb, exists := dp.circuitBreakers[serviceID]; exists {
		return cb
	}

	cb := &circuitBreaker{
		state: cbClosedState,
	}
	dp.circuitBreakers[serviceID] = cb
	return cb
}

// canExecute checks if the circuit breaker allows execution
func (cb *circuitBreaker) canExecute() bool {
	state := atomic.LoadInt32(&cb.state)

	switch state {
	case cbClosedState:
		return true
	case cbOpenState:
		// Check if enough time has passed to try again
		lastFailure := atomic.LoadInt64(&cb.lastFailureTime)
		if time.Since(time.Unix(lastFailure, 0)) > 30*time.Second {
			// Try to transition to half-open
			if atomic.CompareAndSwapInt32(&cb.state, cbOpenState, cbHalfOpenState) {
				return true
			}
			// If CAS failed, check current state again
			return atomic.LoadInt32(&cb.state) == cbHalfOpenState
		}
		return false
	case cbHalfOpenState:
		return true
	default:
		return true
	}
}

// recordSuccess records a successful operation
func (cb *circuitBreaker) recordSuccess() {
	atomic.StoreInt64(&cb.failureCount, 0)
	atomic.StoreInt32(&cb.state, cbClosedState)
}

// recordFailure records a failed operation
func (cb *circuitBreaker) recordFailure() {
	count := atomic.AddInt64(&cb.failureCount, 1)
	atomic.StoreInt64(&cb.lastFailureTime, time.Now().Unix())

	if count >= 3 {
		atomic.StoreInt32(&cb.state, cbOpenState)
	}
}

// isLikelyMovie determines if a download is likely a movie based on title patterns
func (dp *DownloadPoller) isLikelyMovie(item downloadclient.Item) bool {
	if item.Name == "" {
		return false
	}
	lowerTitle := strings.ToLower(item.Name)

	// Look for TV show patterns first - if found, it's not a movie
	if utils.MatchesAnyPattern(lowerTitle, tvPatterns, false) {
		return false
	}

	// Look for movie patterns
	return utils.MatchesAnyPattern(lowerTitle, moviePatterns, false)
}

// isLikelyTVShow determines if a download is likely a TV show based on title patterns
func (dp *DownloadPoller) isLikelyTVShow(item downloadclient.Item) bool {
	if item.Name == "" {
		return false
	}
	lowerTitle := strings.ToLower(item.Name)

	// Look for TV show patterns
	return utils.MatchesAnyPattern(lowerTitle, tvPatterns, false)
}

// extractMovieTitle attempts to extract a clean movie title from a download name
func (dp *DownloadPoller) extractMovieTitle(name string) string {
	// Remove common quality indicators and file extensions
	title := dp.cleanTitle(name)

	// Remove year if present at the end
	title = regexp.MustCompile(`\s*\(\d{4}\)\s*$`).ReplaceAllString(title, "")
	title = regexp.MustCompile(`\s*\d{4}\s*$`).ReplaceAllString(title, "")

	return strings.TrimSpace(title)
}

// extractTVShowTitle attempts to extract a clean TV show title from a download name
func (dp *DownloadPoller) extractTVShowTitle(name string) string {
	// Remove common quality indicators and file extensions
	title := dp.cleanTitle(name)

	// Remove season/episode information
	title = regexp.MustCompile(`\s*S\d{1,2}E\d{1,2}.*$`).ReplaceAllString(title, "")
	title = regexp.MustCompile(`\s*\d{1,2}x\d{1,2}.*$`).ReplaceAllString(title, "")

	return strings.TrimSpace(title)
}

// --- Helper functions for fetching queues and details ---

type radarrQueueItem struct {
	DownloadID string `json:"downloadId"`
	Title      string `json:"title"`
	MovieID    int    `json:"movieId"`
	Size       int64  `json:"size"`
	SizeLeft   int64  `json:"sizeleft"`
	Status     string `json:"status"`
	TimeLeft   string `json:"timeleft"`
}

type sonarrQueueItem struct {
	DownloadID string `json:"downloadId"`
	Title      string `json:"title"`
	SeriesID   int    `json:"seriesId"`
	EpisodeID  int    `json:"episodeId"`
	Size       int64  `json:"size"`
	SizeLeft   int64  `json:"sizeleft"`
	Status     string `json:"status"`
	TimeLeft   string `json:"timeleft"`
}

func fetchRadarrQueue(ctx context.Context, baseURL, apiKey string) ([]radarrQueueItem, error) {
	url := utils.BuildURL(baseURL, "/api/v3/queue", nil)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("X-Api-Key", apiKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if !utils.IsHTTPSuccess(resp.StatusCode) {
		return nil, fmt.Errorf("Radarr queue API request failed: %s", resp.Status)
	}
	var result struct {
		Records []radarrQueueItem `json:"records"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Records, nil
}

func fetchRadarrMovie(ctx context.Context, baseURL, apiKey string, movieID int) (struct {
	TmdbID int
	Title  string
}, error) {
	url := utils.BuildURL(baseURL, fmt.Sprintf("/api/v3/movie/%d", movieID), nil)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("X-Api-Key", apiKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		return struct {
			TmdbID int
			Title  string
		}{}, err
	}
	defer resp.Body.Close()
	if !utils.IsHTTPSuccess(resp.StatusCode) {
		return struct {
			TmdbID int
			Title  string
		}{}, fmt.Errorf("Radarr movie API request failed: %s", resp.Status)
	}
	var movie struct {
		TmdbID int
		Title  string
	}
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		return struct {
			TmdbID int
			Title  string
		}{}, err
	}
	return movie, nil
}

func fetchSonarrQueue(ctx context.Context, baseURL, apiKey string) ([]sonarrQueueItem, error) {
	url := utils.BuildURL(baseURL, "/api/v3/queue", map[string]string{"pageSize": "100"})
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("X-Api-Key", apiKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if !utils.IsHTTPSuccess(resp.StatusCode) {
		return nil, fmt.Errorf("Sonarr queue API request failed: %s", resp.Status)
	}
	var result struct {
		Records []sonarrQueueItem `json:"records"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Records, nil
}

func fetchSonarrSeries(ctx context.Context, baseURL, apiKey string, seriesID int) (struct {
	TmdbID int
	Title  string
}, error) {
	url := utils.BuildURL(baseURL, fmt.Sprintf("/api/v3/series/%d", seriesID), nil)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("X-Api-Key", apiKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		return struct {
			TmdbID int
			Title  string
		}{}, err
	}
	defer resp.Body.Close()
	if !utils.IsHTTPSuccess(resp.StatusCode) {
		return struct {
			TmdbID int
			Title  string
		}{}, fmt.Errorf("Sonarr series API request failed: %s", resp.Status)
	}
	var series struct {
		TmdbID int
		Title  string
	}
	if err := json.NewDecoder(resp.Body).Decode(&series); err != nil {
		return struct {
			TmdbID int
			Title  string
		}{}, err
	}
	return series, nil
}

func fetchSonarrEpisode(ctx context.Context, baseURL, apiKey string, episodeID int) (struct {
	SeasonNumber, EpisodeNumber int
	Title                       string
}, error) {
	url := utils.BuildURL(baseURL, fmt.Sprintf("/api/v3/episode/%d", episodeID), nil)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("X-Api-Key", apiKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		return struct {
			SeasonNumber, EpisodeNumber int
			Title                       string
		}{}, err
	}
	defer resp.Body.Close()
	if !utils.IsHTTPSuccess(resp.StatusCode) {
		return struct {
			SeasonNumber, EpisodeNumber int
			Title                       string
		}{}, fmt.Errorf("Sonarr episode API request failed: %s", resp.Status)
	}
	var episode struct {
		SeasonNumber  int
		EpisodeNumber int
		Title         string
	}
	if err := json.NewDecoder(resp.Body).Decode(&episode); err != nil {
		return struct {
			SeasonNumber, EpisodeNumber int
			Title                       string
		}{}, err
	}
	return episode, nil
}
