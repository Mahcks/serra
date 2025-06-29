package integrations

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
	"time"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/internal/websocket"
	"github.com/mahcks/serra/pkg/downloadclient"
	"github.com/mahcks/serra/pkg/structures"
)

// DownloadPoller is a refactored version that uses the new client architecture
type DownloadPoller struct {
	gctx          global.Context
	ticker        *time.Ticker
	stopChan      chan struct{}
	clientManager *DownloadClientManager

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

	lastCleanupTime time.Time // For improved cleanup scheduling
}

type DownloadPollerOptions struct{}

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
	failureCount    int
	lastFailureTime time.Time
	state           string
	mutex           sync.RWMutex
}

func NewDownloadPoller(gctx global.Context, _ DownloadPollerOptions) (*DownloadPoller, error) {
	dp := &DownloadPoller{
		gctx:          gctx,
		ticker:        time.NewTicker(15 * time.Second),
		stopChan:      make(chan struct{}),
		clientManager: NewDownloadClientManager(),
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
		circuitBreakers: make(map[string]*circuitBreaker),
		lastCleanupTime: time.Now(),
	}

	// Initialize download clients
	if err := dp.initializeClients(); err != nil {
		return nil, err
	}

	// Start immediately on creation
	dp.pollCombined()

	// Start the polling goroutine
	go dp.startPolling()

	return dp, nil
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
	slog.Debug("Starting poll cycle (multi-instance, queue-first)")

	ctx := context.Background()

	// 1. Fetch all downloads from all clients (for matching)
	allClientDownloads, err := dp.clientManager.GetAllDownloads(ctx)
	if err != nil {
		slog.Error("Failed to get downloads from clients", "error", err)
		dp.errorCount++
		return
	}
	slog.Debug("Fetched downloads from all clients", "count", len(allClientDownloads))

	// Build a map for fast lookup by hash and by name
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

	var allEnrichedDownloads []Download

	// 2. For each Radarr instance, fetch queue and enrich
	radarrInstances, err := dp.gctx.Crate().Sqlite.Query().GetArrServiceByType(ctx, "radarr")
	if err != nil {
		slog.Error("Failed to fetch Radarr instances", "error", err)
	} else {
		for _, radarr := range radarrInstances {
			slog.Info("Polling Radarr queue", "name", radarr.Name, "url", radarr.BaseUrl)
			queue, err := fetchRadarrQueue(radarr.BaseUrl, radarr.ApiKey)
			if err != nil {
				slog.Error("Failed to fetch Radarr queue", "name", radarr.Name, "error", err)
				continue
			}
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
				movie, err := fetchRadarrMovie(radarr.BaseUrl, radarr.ApiKey, item.MovieID)
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
					TmdbID:       ptrInt64(int64(movie.TmdbID)),
					Progress:     progress,
					TimeLeft:     ptrString(timeLeft),
					Status:       ptrString(status),
					Hash:         ptrString(hash),
				})
			}
		}
	}

	// 3. For each Sonarr instance, fetch queue and enrich
	sonarrInstances, err := dp.gctx.Crate().Sqlite.Query().GetArrServiceByType(ctx, "sonarr")
	if err != nil {
		slog.Error("Failed to fetch Sonarr instances", "error", err)
	} else {
		for _, sonarr := range sonarrInstances {
			slog.Info("Polling Sonarr queue", "name", sonarr.Name, "url", sonarr.BaseUrl)
			queue, err := fetchSonarrQueue(sonarr.BaseUrl, sonarr.ApiKey)
			if err != nil {
				slog.Error("Failed to fetch Sonarr queue", "name", sonarr.Name, "error", err)
				continue
			}
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
				series, err := fetchSonarrSeries(sonarr.BaseUrl, sonarr.ApiKey, item.SeriesID)
				if err != nil {
					slog.Info("Failed to fetch Sonarr series details", "seriesID", item.SeriesID, "error", err)
					continue
				}
				episode, err := fetchSonarrEpisode(sonarr.BaseUrl, sonarr.ApiKey, item.EpisodeID)
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
					TmdbID:       ptrInt64(int64(series.TmdbID)),
					Progress:     progress,
					TimeLeft:     ptrString(timeLeft),
					Status:       ptrString(status),
					Hash:         ptrString(hash),
				})
			}
		}
	}

	// 4. Store downloads in database
	dp.storeDownloads(allEnrichedDownloads)

	// 5. Broadcast via WebSocket
	if len(allEnrichedDownloads) > 0 {
		var batch []structures.DownloadProgressPayload
		for _, d := range allEnrichedDownloads {
			batch = append(batch, structures.DownloadProgressPayload{
				ID:           d.ID,
				Title:        d.Title,
				TorrentTitle: d.TorrentTitle,
				Source:       d.Source,
				TMDBID:       d.TmdbID,
				TvDBID:       nil, // Not used in current implementation
				Hash:         derefString(d.Hash),
				Progress:     d.Progress,
				TimeLeft:     derefString(d.TimeLeft),
				Status:       derefString(d.Status),
				LastUpdated:  time.Now().Format(time.RFC3339),
			})
		}
		websocket.BroadcastToAll(structures.OpcodeDownloadProgressBatch, structures.DownloadProgressBatchPayload{
			Downloads: batch,
		})
		slog.Info("Found active downloads", "count", len(allEnrichedDownloads))
	}

	// Clean up completed downloads from database
	dp.cleanupCompletedDownloads(allEnrichedDownloads)

	// Clean up old missing downloads every 25 minutes
	if time.Since(dp.lastCleanupTime) > 25*time.Minute {
		dp.cleanupOldMissingDownloads()
		dp.cleanupCount++
		dp.lastCleanupTime = time.Now()
	}

	// Update metrics
	dp.lastPollTime = time.Now()
	dp.pollCount++
	dp.downloadsFound += int64(len(allEnrichedDownloads))

	// Adaptive polling: adjust interval based on activity
	dp.adjustPollInterval(len(allEnrichedDownloads))

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
		slog.Debug("Matched with Radarr", "id", item.ID, "title", matchedDownload.Title)
		return matchedDownload
	}

	matchedDownload = dp.matchWithSonarr(item)
	if matchedDownload != nil {
		slog.Debug("Matched with Sonarr", "id", item.ID, "title", matchedDownload.Title)
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
		TimeLeft:     ptrString(item.TimeLeft),
		Status:       ptrString(item.Status),
		Hash:         ptrString(item.Hash),
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
	lowerTitle := strings.ToLower(item.Name)

	// Look for TV show patterns (season/episode indicators)
	tvPatterns := []string{
		"s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8", "s9", "s10", "s11", "s12", "s13", "s14", "s15", "s16", "s17", "s18", "s19", "s20",
		"e0", "e1", "e2", "e3", "e4", "e5", "e6", "e7", "e8", "e9", "e10", "e11", "e12", "e13", "e14", "e15", "e16", "e17", "e18", "e19", "e20",
		"season", "episode", "s01e", "s02e", "s03e", "s04e", "s05e", "s06e", "s07e", "s08e", "s09e", "s10e", "s11e", "s12e", "s13e", "s14e", "s15e", "s16e", "s17e", "s18e", "s19e", "s20e",
	}

	for _, pattern := range tvPatterns {
		if strings.Contains(lowerTitle, pattern) {
			return "sonarr"
		}
	}

	// Look for movie patterns (year indicators, movie-specific terms)
	moviePatterns := []string{
		"1080p", "720p", "4k", "2160p", "bluray", "web-dl", "hdtv", "dvdrip", "brrip",
		"x264", "x265", "h264", "h265", "aac", "ac3", "dts",
	}

	// If it has movie patterns but no TV patterns, likely a movie
	hasMoviePattern := false
	for _, pattern := range moviePatterns {
		if strings.Contains(lowerTitle, pattern) {
			hasMoviePattern = true
			break
		}
	}

	if hasMoviePattern {
		// Check if it also has TV patterns
		hasTVPattern := false
		for _, pattern := range tvPatterns {
			if strings.Contains(lowerTitle, pattern) {
				hasTVPattern = true
				break
			}
		}

		if !hasTVPattern {
			return "radarr"
		}
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
					TmdbID:       ptrInt64(int64(movie.TmdbID)),
					Progress:     item.Progress,
					TimeLeft:     ptrString(item.TimeLeft),
					Status:       ptrString(item.Status),
					Hash:         ptrString(item.Hash),
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
			TimeLeft:     ptrString(item.TimeLeft),
			Status:       ptrString(item.Status),
			Hash:         ptrString(item.Hash),
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
					TmdbID:       ptrInt64(int64(series.TmdbID)),
					Progress:     item.Progress,
					TimeLeft:     ptrString(item.TimeLeft),
					Status:       ptrString(item.Status),
					Hash:         ptrString(item.Hash),
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
						TmdbID:       ptrInt64(int64(series.TmdbID)),
						Progress:     item.Progress,
						TimeLeft:     ptrString(item.TimeLeft),
						Status:       ptrString(item.Status),
						Hash:         ptrString(item.Hash),
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
			TimeLeft:     ptrString(item.TimeLeft),
			Status:       ptrString(item.Status),
			Hash:         ptrString(item.Hash),
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
	downloadTitle := strings.ToLower(item.Name)

	// Look for season/episode patterns like S01E02, 1x02, etc.
	seasonPattern := fmt.Sprintf("s%02d", episode.SeasonNumber)
	episodePattern := fmt.Sprintf("e%02d", episode.EpisodeNumber)

	if strings.Contains(downloadTitle, seasonPattern) && strings.Contains(downloadTitle, episodePattern) {
		slog.Debug("Episode match found by season/episode pattern",
			"download", item.Name,
			"episode", fmt.Sprintf("S%02dE%02d", episode.SeasonNumber, episode.EpisodeNumber))
		return true
	}

	// Alternative patterns like 1x02, 01x02
	altSeasonPattern := fmt.Sprintf("%dx", episode.SeasonNumber)
	altEpisodePattern := fmt.Sprintf("x%02d", episode.EpisodeNumber)

	if strings.Contains(downloadTitle, altSeasonPattern) && strings.Contains(downloadTitle, altEpisodePattern) {
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
	for _, ext := range extensions {
		title = strings.ReplaceAll(title, ext, "")
	}

	// Remove quality indicators
	qualities := []string{"1080p", "720p", "480p", "2160p", "4k", "hdtv", "web-dl", "bluray", "dvdrip", "hddvd", "brrip"}
	for _, quality := range qualities {
		title = strings.ReplaceAll(title, quality, "")
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
		slog.Debug("Storing download",
			"id", download.ID,
			"title", download.Title,
			"progress", download.Progress,
			"status", derefString(download.Status),
			"timeLeft", derefString(download.TimeLeft),
			"source", download.Source,
			"tmdb_id", download.TmdbID)

		// Handle nullable fields properly
		var tmdbID sql.NullInt64
		if download.TmdbID != nil {
			tmdbID = sql.NullInt64{Int64: *download.TmdbID, Valid: true}
			slog.Info("Storing TMDb ID", "id", download.ID, "tmdb_id", *download.TmdbID)
		}

		var hash sql.NullString
		if download.Hash != nil {
			hash = sql.NullString{String: *download.Hash, Valid: true}
		}

		var timeLeft sql.NullString
		if download.TimeLeft != nil {
			timeLeft = sql.NullString{String: *download.TimeLeft, Valid: true}
		}

		var status sql.NullString
		if download.Status != nil {
			status = sql.NullString{String: *download.Status, Valid: true}
		}

		err := dp.gctx.Crate().Sqlite.Query().UpsertDownloadQueue(context.Background(), repository.UpsertDownloadQueueParams{
			ID:           download.ID,
			Title:        download.Title,
			TorrentTitle: download.TorrentTitle,
			Source:       download.Source,
			TmdbID:       tmdbID,
			Hash:         hash,
			Progress:     sql.NullFloat64{Float64: download.Progress, Valid: true},
			TimeLeft:     timeLeft,
			Status:       status,
		})
		if err != nil {
			slog.Error("Failed to upsert download", "id", download.ID, "error", err)
		} else {
			slog.Debug("Successfully stored download", "id", download.ID, "title", download.Title)
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
			slog.Info("Cleaned up old missing download", "id", download.ID, "title", download.Title, "age", time.Since(download.LastUpdated.Time))
		}
	}

	if len(downloads) > 0 {
		slog.Info("Cleaned up old missing downloads", "count", len(downloads))
	}
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
	moviesRaw, err := dp.fetchRadarrMovies(baseURL, apiKey)
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
func (dp *DownloadPoller) fetchRadarrMovies(baseURL, apiKey string) (map[int]struct {
	TmdbID      int    `json:"tmdbId"`
	Title       string `json:"title"`
	LastUpdated time.Time
}, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	url := fmt.Sprintf("%s/api/v3/movie?apikey=%s", baseURL, apiKey)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
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
	series, episodes, err := dp.fetchSonarrData(baseURL, apiKey)
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
func (dp *DownloadPoller) fetchSonarrData(baseURL, apiKey string) (map[int]struct {
	TmdbID int    `json:"tmdbId"`
	Title  string `json:"title"`
}, map[int]sonarrEpisode, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	seriesURL := fmt.Sprintf("%s/api/v3/series?apikey=%s", baseURL, apiKey)
	seriesReq, err := http.NewRequest("GET", seriesURL, nil)
	if err != nil {
		return nil, nil, err
	}
	seriesResp, err := client.Do(seriesReq)
	if err != nil {
		return nil, nil, err
	}
	defer seriesResp.Body.Close()
	if seriesResp.StatusCode != http.StatusOK {
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
		episodesURL := fmt.Sprintf("%s/api/v3/episode?seriesId=%d&apikey=%s", baseURL, s.ID, apiKey)
		episodesReq, err := http.NewRequest("GET", episodesURL, nil)
		if err != nil {
			continue // Skip this series if we can't create request
		}
		episodesResp, err := client.Do(episodesReq)
		if err != nil {
			continue // Skip this series if request fails
		}
		if episodesResp.StatusCode != http.StatusOK {
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
		state: cbClosed,
	}
	dp.circuitBreakers[serviceID] = cb
	return cb
}

// canExecute checks if the circuit breaker allows execution
func (cb *circuitBreaker) canExecute() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.state {
	case cbClosed:
		return true
	case cbOpen:
		// Check if enough time has passed to try again
		if time.Since(cb.lastFailureTime) > 30*time.Second {
			cb.mutex.RUnlock()
			cb.mutex.Lock()
			cb.state = cbHalfOpen
			cb.mutex.Unlock()
			cb.mutex.RLock()
			return true
		}
		return false
	case cbHalfOpen:
		return true
	default:
		return true
	}
}

// recordSuccess records a successful operation
func (cb *circuitBreaker) recordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount = 0
	cb.state = cbClosed
}

// recordFailure records a failed operation
func (cb *circuitBreaker) recordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.failureCount >= 3 {
		cb.state = cbOpen
	}
}

// Helper function to convert string to pointer
func ptrString(val string) *string {
	return &val
}

// Helper function to convert int64 to pointer
func ptrInt64(val int64) *int64 {
	return &val
}

// Helper to dereference *string safely
func derefString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// isLikelyMovie determines if a download is likely a movie based on title patterns
func (dp *DownloadPoller) isLikelyMovie(item downloadclient.Item) bool {
	lowerTitle := strings.ToLower(item.Name)

	// Look for TV show patterns first - if found, it's not a movie
	tvPatterns := []string{"s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8", "s9", "s10", "s11", "s12", "s13", "s14", "s15", "s16", "s17", "s18", "s19", "s20",
		"e0", "e1", "e2", "e3", "e4", "e5", "e6", "e7", "e8", "e9", "e10", "e11", "e12", "e13", "e14", "e15", "e16", "e17", "e18", "e19", "e20",
		"season", "episode", "s01e", "s02e", "s03e", "s04e", "s05e", "s06e", "s07e", "s08e", "s09e", "s10e", "s11e", "s12e", "s13e", "s14e", "s15e", "s16e", "s17e", "s18e", "s19e", "s20e"}

	for _, pattern := range tvPatterns {
		if strings.Contains(lowerTitle, pattern) {
			return false
		}
	}

	// Look for movie patterns
	moviePatterns := []string{"1080p", "720p", "4k", "2160p", "bluray", "web-dl", "hdtv", "dvdrip", "brrip"}
	for _, pattern := range moviePatterns {
		if strings.Contains(lowerTitle, pattern) {
			return true
		}
	}

	return false
}

// isLikelyTVShow determines if a download is likely a TV show based on title patterns
func (dp *DownloadPoller) isLikelyTVShow(item downloadclient.Item) bool {
	lowerTitle := strings.ToLower(item.Name)

	// Look for TV show patterns
	tvPatterns := []string{"s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8", "s9", "s10", "s11", "s12", "s13", "s14", "s15", "s16", "s17", "s18", "s19", "s20",
		"e0", "e1", "e2", "e3", "e4", "e5", "e6", "e7", "e8", "e9", "e10", "e11", "e12", "e13", "e14", "e15", "e16", "e17", "e18", "e19", "e20",
		"season", "episode", "s01e", "s02e", "s03e", "s04e", "s05e", "s06e", "s07e", "s08e", "s09e", "s10e", "s11e", "s12e", "s13e", "s14e", "s15e", "s16e", "s17e", "s18e", "s19e", "s20e"}

	for _, pattern := range tvPatterns {
		if strings.Contains(lowerTitle, pattern) {
			return true
		}
	}

	return false
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

func fetchRadarrQueue(baseURL, apiKey string) ([]radarrQueueItem, error) {
	url := fmt.Sprintf("%s/api/v3/queue", baseURL)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Api-Key", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
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

func fetchRadarrMovie(baseURL, apiKey string, movieID int) (struct {
	TmdbID int
	Title  string
}, error) {
	url := fmt.Sprintf("%s/api/v3/movie/%d", baseURL, movieID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Api-Key", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return struct {
			TmdbID int
			Title  string
		}{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
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

func fetchSonarrQueue(baseURL, apiKey string) ([]sonarrQueueItem, error) {
	url := fmt.Sprintf("%s/api/v3/queue?pageSize=100", baseURL)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Api-Key", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
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

func fetchSonarrSeries(baseURL, apiKey string, seriesID int) (struct {
	TmdbID int
	Title  string
}, error) {
	url := fmt.Sprintf("%s/api/v3/series/%d", baseURL, seriesID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Api-Key", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return struct {
			TmdbID int
			Title  string
		}{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
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

func fetchSonarrEpisode(baseURL, apiKey string, episodeID int) (struct {
	SeasonNumber, EpisodeNumber int
	Title                       string
}, error) {
	url := fmt.Sprintf("%s/api/v3/episode/%d", baseURL, episodeID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Api-Key", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return struct {
			SeasonNumber, EpisodeNumber int
			Title                       string
		}{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
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
