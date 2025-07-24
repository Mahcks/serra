package sonarr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/pkg/structures"
)

type Service interface {
	GetUpcomingItems(ctx context.Context) ([]structures.CalendarItem, error)
	AddSeries(ctx context.Context, tmdbID int64, qualityProfileID int, rootFolderPath string) (*AddSeriesResponse, error)
	AddSeriesWithSeasons(ctx context.Context, tmdbID int64, qualityProfileID int, rootFolderPath string, seasons []int) (*AddSeriesResponse, error)
	GetSeriesByTMDBID(ctx context.Context, tmdbID int64) (*SeriesResponse, error)
	SearchSeries(ctx context.Context, seriesID int) error
}

type AddSeriesResponse struct {
	ID               int    `json:"id"`
	Title            string `json:"title"`
	TmdbID           int64  `json:"tmdbId"`
	QualityProfileID int    `json:"qualityProfileId"`
	RootFolderPath   string `json:"rootFolderPath"`
	Monitored        bool   `json:"monitored"`
	Added            string `json:"added"`
}

type SeriesResponse struct {
	ID               int    `json:"id"`
	Title            string `json:"title"`
	TmdbID           int64  `json:"tmdbId"`
	QualityProfileID int    `json:"qualityProfileId"`
	RootFolderPath   string `json:"rootFolderPath"`
	Monitored        bool   `json:"monitored"`
	Status           string `json:"status"`
	Statistics       struct {
		EpisodeFileCount int `json:"episodeFileCount"`
		EpisodeCount     int `json:"episodeCount"`
		TotalEpisodeCount int `json:"totalEpisodeCount"`
		PercentOfEpisodes float64 `json:"percentOfEpisodes"`
	} `json:"statistics"`
}

type AddSeriesRequest struct {
	Title            string `json:"title"`
	TmdbID           int64  `json:"tmdbId"`
	QualityProfileID int    `json:"qualityProfileId"`
	RootFolderPath   string `json:"rootFolderPath"`
	Monitored        bool   `json:"monitored"`
	SearchForMissingEpisodes bool `json:"searchForMissingEpisodes"`
	MonitorType      string `json:"monitorType"`
	Seasons          []SeasonRequest `json:"seasons,omitempty"`
}

type SeasonRequest struct {
	SeasonNumber int  `json:"seasonNumber"`
	Monitored    bool `json:"monitored"`
}

type sonarrService struct {
	repo   *repository.Queries
	client *http.Client
}

func New(repo *repository.Queries) Service {
	return &sonarrService{
		repo:   repo,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (ss *sonarrService) GetUpcomingItems(ctx context.Context) ([]structures.CalendarItem, error) {
	instances, err := ss.repo.GetArrServiceByType(ctx, structures.ProviderSonarr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sonarr instances: %w", err)
	}

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		allItems []structures.CalendarItem
		errs     []error
	)

	for _, inst := range instances {
		inst := inst // capture range variable
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Step 1: Get all series data (so we can map seriesId → title)
			seriesURL := fmt.Sprintf("%s/api/v3/series?apikey=%s", inst.BaseUrl, inst.ApiKey)
			seriesReq, err := http.NewRequestWithContext(ctx, "GET", seriesURL, nil)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("instance %s: failed to create series request: %w", inst.Name, err))
				mu.Unlock()
				return
			}
			seriesResp, err := ss.client.Do(seriesReq)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("instance %s: failed to fetch Sonarr series: %w", inst.Name, err))
				mu.Unlock()
				return
			}
			defer seriesResp.Body.Close()

			var seriesList []struct {
				ID     int    `json:"id"`
				Title  string `json:"title"`
				TmdbID int64  `json:"tmdbId"`
			}
			if err := json.NewDecoder(seriesResp.Body).Decode(&seriesList); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("instance %s: failed to decode Sonarr series response: %w", inst.Name, err))
				mu.Unlock()
				return
			}

			// Build a quick lookup map
			seriesMap := make(map[int]string)
			tmdbMap := make(map[int]int64)
			for _, s := range seriesList {
				seriesMap[s.ID] = s.Title
				tmdbMap[s.ID] = s.TmdbID
			}

			// Step 2: Get all wanted episodes with paging
			page := 1
			pageSize := 100

			for {
				url := fmt.Sprintf("%s/api/v3/wanted/missing?apikey=%s&page=%d&pageSize=%d", inst.BaseUrl, inst.ApiKey, page, pageSize)
				req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				if err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("instance %s: failed to create request: %w", inst.Name, err))
					mu.Unlock()
					return
				}
				resp, err := ss.client.Do(req)
				if err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("instance %s: failed to contact Sonarr: %w", inst.Name, err))
					mu.Unlock()
					return
				}
				defer resp.Body.Close()

				var data struct {
					Page         int `json:"page"`
					PageSize     int `json:"pageSize"`
					TotalRecords int `json:"totalRecords"`
					Records      []struct {
						SeriesID      int       `json:"seriesId"`
						ID            int       `json:"id"`
						Title         string    `json:"title"`
						SeasonNumber  int       `json:"seasonNumber"`
						EpisodeNumber int       `json:"episodeNumber"`
						AirDateUtc    time.Time `json:"airDateUtc"`
						HasFile       bool      `json:"hasFile"`
					} `json:"records"`
				}

				if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("instance %s: failed to decode Sonarr response: %w", inst.Name, err))
					mu.Unlock()
					return
				}

				var items []structures.CalendarItem
				for _, r := range data.Records {
					if r.HasFile || r.AirDateUtc.IsZero() {
						continue
					}
					if r.AirDateUtc.Before(time.Now()) {
						continue // only keep future episodes
					}

					seriesTitle, ok := seriesMap[r.SeriesID]
					if !ok {
						seriesTitle = "Unknown Series"
					}

					tmdbID, _ := tmdbMap[r.SeriesID] // Default to 0 if not found

					fullTitle := fmt.Sprintf("%s S%02dE%02d - %s",
						seriesTitle, r.SeasonNumber, r.EpisodeNumber, r.Title)

					items = append(items, structures.CalendarItem{
						Title:       fullTitle,
						Source:      structures.ProviderSonarr,
						ReleaseDate: r.AirDateUtc,
						TmdbID:      tmdbID,
					})
				}

				mu.Lock()
				allItems = append(allItems, items...)
				mu.Unlock()

				if (data.Page * data.PageSize) >= data.TotalRecords {
					break
				}
				page++
			}
		}()
	}

	wg.Wait()

	if len(errs) > 0 {
		return allItems, fmt.Errorf("errors occurred: %v", errs)
	}

	return allItems, nil
}

// AddSeries adds a TV series to Sonarr
func (ss *sonarrService) AddSeries(ctx context.Context, tmdbID int64, qualityProfileID int, rootFolderPath string) (*AddSeriesResponse, error) {
	instances, err := ss.repo.GetArrServiceByType(ctx, structures.ProviderSonarr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sonarr instances: %w", err)
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("no Sonarr instances configured")
	}

	// Use the first available instance
	instance := instances[0]

	// First, check if the series already exists
	existingSeries, _ := ss.GetSeriesByTMDBID(ctx, tmdbID)
	if existingSeries != nil {
		return &AddSeriesResponse{
			ID:               existingSeries.ID,
			Title:            existingSeries.Title,
			TmdbID:           existingSeries.TmdbID,
			QualityProfileID: existingSeries.QualityProfileID,
			RootFolderPath:   existingSeries.RootFolderPath,
			Monitored:        existingSeries.Monitored,
		}, nil
	}

	// Prepare the request
	addRequest := AddSeriesRequest{
		TmdbID:                   tmdbID,
		QualityProfileID:         qualityProfileID,
		RootFolderPath:           rootFolderPath,
		Monitored:                true,
		SearchForMissingEpisodes: true,
		MonitorType:             "all", // Monitor all episodes
	}

	requestBody, err := json.Marshal(addRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v3/series?apikey=%s", instance.BaseUrl, instance.ApiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ss.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to contact Sonarr: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Sonarr returned status %d", resp.StatusCode)
	}

	var response AddSeriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode Sonarr response: %w", err)
	}

	slog.Info("Series added to Sonarr successfully", 
		"seriesID", response.ID,
		"title", response.Title,
		"monitored", response.Monitored,
		"root_folder", response.RootFolderPath)

	// Trigger search for the newly added series
	if err := ss.SearchSeries(ctx, response.ID); err != nil {
		slog.Warn("Failed to trigger automatic search for series", 
			"seriesID", response.ID, 
			"title", response.Title, 
			"error", err)
		// Don't fail the entire operation if search trigger fails
	}

	return &response, nil
}

// AddSeriesWithSeasons adds a TV series to Sonarr with specific season monitoring
func (ss *sonarrService) AddSeriesWithSeasons(ctx context.Context, tmdbID int64, qualityProfileID int, rootFolderPath string, seasons []int) (*AddSeriesResponse, error) {
	instances, err := ss.repo.GetArrServiceByType(ctx, structures.ProviderSonarr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sonarr instances: %w", err)
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("no Sonarr instances configured")
	}

	// Use the first available instance
	instance := instances[0]

	// First, check if the series already exists
	existingSeries, _ := ss.GetSeriesByTMDBID(ctx, tmdbID)
	if existingSeries != nil {
		// Series already exists - we need to update season monitoring
		// For now, return the existing series (could be enhanced to update monitoring)
		return &AddSeriesResponse{
			ID:               existingSeries.ID,
			Title:            existingSeries.Title,
			TmdbID:           existingSeries.TmdbID,
			QualityProfileID: existingSeries.QualityProfileID,
			RootFolderPath:   existingSeries.RootFolderPath,
			Monitored:        existingSeries.Monitored,
		}, nil
	}

	// Build season monitoring information
	var seasonRequests []SeasonRequest
	if len(seasons) > 0 {
		// Monitor only requested seasons
		for _, seasonNum := range seasons {
			seasonRequests = append(seasonRequests, SeasonRequest{
				SeasonNumber: seasonNum,
				Monitored:    true,
			})
		}
	}

	// Determine monitor type based on seasons
	monitorType := "all"
	if len(seasons) > 0 {
		monitorType = "none" // We'll specify seasons manually
	}

	// Prepare the request
	addRequest := AddSeriesRequest{
		TmdbID:                   tmdbID,
		QualityProfileID:         qualityProfileID,
		RootFolderPath:           rootFolderPath,
		Monitored:                true,
		SearchForMissingEpisodes: true,
		MonitorType:              monitorType,
		Seasons:                  seasonRequests,
	}

	requestBody, err := json.Marshal(addRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v3/series?apikey=%s", instance.BaseUrl, instance.ApiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ss.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to contact Sonarr: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Sonarr returned status %d", resp.StatusCode)
	}

	var response AddSeriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode Sonarr response: %w", err)
	}

	slog.Info("Series with seasons added to Sonarr successfully", 
		"seriesID", response.ID,
		"title", response.Title,
		"monitored", response.Monitored,
		"root_folder", response.RootFolderPath)

	// Trigger search for the newly added series
	if err := ss.SearchSeries(ctx, response.ID); err != nil {
		slog.Warn("Failed to trigger automatic search for series with seasons", 
			"seriesID", response.ID, 
			"title", response.Title, 
			"error", err)
		// Don't fail the entire operation if search trigger fails
	}

	return &response, nil
}

// GetSeriesByTMDBID retrieves a TV series from Sonarr by TMDB ID
func (ss *sonarrService) GetSeriesByTMDBID(ctx context.Context, tmdbID int64) (*SeriesResponse, error) {
	instances, err := ss.repo.GetArrServiceByType(ctx, structures.ProviderSonarr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sonarr instances: %w", err)
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("no Sonarr instances configured")
	}

	instance := instances[0]

	url := fmt.Sprintf("%s/api/v3/series?apikey=%s&tmdbId=%d", instance.BaseUrl, instance.ApiKey, tmdbID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := ss.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to contact Sonarr: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Series not found
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Sonarr returned status %d", resp.StatusCode)
	}

	var series []SeriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&series); err != nil {
		return nil, fmt.Errorf("failed to decode Sonarr response: %w", err)
	}

	if len(series) == 0 {
		return nil, nil // Series not found
	}

	return &series[0], nil
}

// SearchSeries triggers a search for missing episodes for a specific series in Sonarr
func (ss *sonarrService) SearchSeries(ctx context.Context, seriesID int) error {
	instances, err := ss.repo.GetArrServiceByType(ctx, structures.ProviderSonarr.String())
	if err != nil {
		return fmt.Errorf("failed to fetch sonarr instances: %w", err)
	}

	if len(instances) == 0 {
		return fmt.Errorf("no Sonarr instances configured")
	}

	// Try all instances until one succeeds
	for _, instance := range instances {
		err := ss.searchSeriesOnInstance(ctx, instance, seriesID)
		if err == nil {
			slog.Info("Series search triggered successfully", "seriesID", seriesID, "instance", instance.Name)
			return nil
		}
		slog.Warn("Failed to trigger search on instance", "instance", instance.Name, "error", err)
	}

	return fmt.Errorf("failed to trigger search on any Sonarr instance")
}

func (ss *sonarrService) searchSeriesOnInstance(ctx context.Context, instance repository.ArrService, seriesID int) error {
	// Sonarr command API to trigger series search
	searchCommand := map[string]interface{}{
		"name":      "SeriesSearch",
		"seriesIds": []int{seriesID},
	}

	requestBody, err := json.Marshal(searchCommand)
	if err != nil {
		return fmt.Errorf("failed to marshal search command: %w", err)
	}

	url := fmt.Sprintf("%s/api/v3/command?apikey=%s", instance.BaseUrl, instance.ApiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ss.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to contact Sonarr: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("Sonarr search command failed", 
			"seriesID", seriesID, 
			"status", resp.StatusCode, 
			"response", string(body),
			"instance", instance.Name)
		return fmt.Errorf("Sonarr returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
