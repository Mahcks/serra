package radarr

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
	GetCalendarItems(ctx context.Context) ([]structures.CalendarItem, error)
	AddMovie(ctx context.Context, tmdbID int64, qualityProfileID int, rootFolderPath string, minimumAvailability string) (*AddMovieResponse, error)
	GetMovieByTMDBID(ctx context.Context, tmdbID int64) (*MovieResponse, error)
	SearchMovie(ctx context.Context, movieID int) error
}

type AddMovieResponse struct {
	ID                  int    `json:"id"`
	Title               string `json:"title"`
	TmdbID              int64  `json:"tmdbId"`
	QualityProfileID    int    `json:"qualityProfileId"`
	RootFolderPath      string `json:"rootFolderPath"`
	MinimumAvailability string `json:"minimumAvailability"`
	Monitored           bool   `json:"monitored"`
	Added               string `json:"added"`
}

type MovieResponse struct {
	ID                  int    `json:"id"`
	Title               string `json:"title"`
	TmdbID              int64  `json:"tmdbId"`
	Downloaded          bool   `json:"downloaded"`
	HasFile             bool   `json:"hasFile"`
	Status              string `json:"status"`
	QualityProfileID    int    `json:"qualityProfileId"`
	RootFolderPath      string `json:"rootFolderPath"`
	MinimumAvailability string `json:"minimumAvailability"`
	Monitored           bool   `json:"monitored"`
}

type AddMovieRequest struct {
	Title               string `json:"title"`
	TmdbID              int64  `json:"tmdbId"`
	QualityProfileID    int    `json:"qualityProfileId"`
	RootFolderPath      string `json:"rootFolderPath"`
	MinimumAvailability string `json:"minimumAvailability"`
	Monitored           bool   `json:"monitored"`
	SearchForMovie      bool   `json:"searchForMovie"`
}

type radarrService struct {
	repo   *repository.Queries
	client *http.Client
}

func New(repo *repository.Queries) Service {
	return &radarrService{
		repo:   repo,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (rs *radarrService) GetCalendarItems(ctx context.Context) ([]structures.CalendarItem, error) {
	instances, err := rs.repo.GetArrServiceByType(ctx, structures.ProviderRadarr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch radarr instances: %w", err)
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
				resp, err := rs.client.Do(req)
				if err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("instance %s: failed to contact Radarr: %w", inst.Name, err))
					mu.Unlock()
					return
				}
				defer resp.Body.Close()

				var data struct {
					Page         int `json:"page"`
					PageSize     int `json:"pageSize"`
					TotalRecords int `json:"totalRecords"`
					Records      []struct {
						ID             int    `json:"id"`
						Title          string `json:"title"`
						TmdbID         int64  `json:"tmdbId"`
						HasFile        bool   `json:"hasFile"`
						DigitalRelease string `json:"digitalRelease"`
					} `json:"records"`
				}

				if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("instance %s: failed to decode Radarr response: %w", inst.Name, err))
					mu.Unlock()
					return
				}

				var items []structures.CalendarItem
				for _, r := range data.Records {
					if r.HasFile || r.DigitalRelease == "" {
						continue
					}
					releaseTime, err := time.Parse(time.RFC3339, r.DigitalRelease)
					if err != nil {
						continue
					}
					if releaseTime.Before(time.Now()) {
						continue
					}
					items = append(items, structures.CalendarItem{
						Title:       r.Title,
						Source:      structures.ProviderRadarr,
						ReleaseDate: releaseTime,
						TmdbID:      r.TmdbID,
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

// AddMovie adds a movie to Radarr
func (rs *radarrService) AddMovie(ctx context.Context, tmdbID int64, qualityProfileID int, rootFolderPath string, minimumAvailability string) (*AddMovieResponse, error) {
	instances, err := rs.repo.GetArrServiceByType(ctx, structures.ProviderRadarr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch radarr instances: %w", err)
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("no Radarr instances configured")
	}

	// Use the first available instance (or the first 4K if it's a 4K request)
	instance := instances[0]

	// First, check if the movie already exists
	existingMovie, _ := rs.GetMovieByTMDBID(ctx, tmdbID)
	if existingMovie != nil {
		return &AddMovieResponse{
			ID:                  existingMovie.ID,
			Title:               existingMovie.Title,
			TmdbID:              existingMovie.TmdbID,
			QualityProfileID:    existingMovie.QualityProfileID,
			RootFolderPath:      existingMovie.RootFolderPath,
			MinimumAvailability: existingMovie.MinimumAvailability,
			Monitored:           existingMovie.Monitored,
		}, nil
	}

	// Prepare the request
	addRequest := AddMovieRequest{
		TmdbID:              tmdbID,
		QualityProfileID:    qualityProfileID,
		RootFolderPath:      rootFolderPath,
		MinimumAvailability: minimumAvailability,
		Monitored:           true,
		SearchForMovie:      true,
	}

	requestBody, err := json.Marshal(addRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v3/movie?apikey=%s", instance.BaseUrl, instance.ApiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := rs.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to contact Radarr: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Radarr returned status %d", resp.StatusCode)
	}

	var response AddMovieResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode Radarr response: %w", err)
	}

	slog.Info("Movie added to Radarr successfully", 
		"radarr_id", response.ID,
		"title", response.Title,
		"monitored", response.Monitored,
		"root_folder", response.RootFolderPath)

	// Trigger search for the newly added movie
	if err := rs.SearchMovie(ctx, response.ID); err != nil {
		slog.Warn("Failed to trigger automatic search for movie", 
			"movieID", response.ID, 
			"title", response.Title, 
			"error", err)
		// Don't fail the entire operation if search trigger fails
	}

	return &response, nil
}

// GetMovieByTMDBID retrieves a movie from Radarr by TMDB ID
func (rs *radarrService) GetMovieByTMDBID(ctx context.Context, tmdbID int64) (*MovieResponse, error) {
	instances, err := rs.repo.GetArrServiceByType(ctx, structures.ProviderRadarr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch radarr instances: %w", err)
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("no Radarr instances configured")
	}

	instance := instances[0]

	url := fmt.Sprintf("%s/api/v3/movie?apikey=%s&tmdbId=%d", instance.BaseUrl, instance.ApiKey, tmdbID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := rs.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to contact Radarr: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Movie not found
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Radarr returned status %d", resp.StatusCode)
	}

	var movies []MovieResponse
	if err := json.NewDecoder(resp.Body).Decode(&movies); err != nil {
		return nil, fmt.Errorf("failed to decode Radarr response: %w", err)
	}

	if len(movies) == 0 {
		return nil, nil // Movie not found
	}

	return &movies[0], nil
}

// SearchMovie triggers a search for a specific movie in Radarr
func (rs *radarrService) SearchMovie(ctx context.Context, movieID int) error {
	instances, err := rs.repo.GetArrServiceByType(ctx, structures.ProviderRadarr.String())
	if err != nil {
		return fmt.Errorf("failed to fetch radarr instances: %w", err)
	}

	if len(instances) == 0 {
		return fmt.Errorf("no Radarr instances configured")
	}

	// Try all instances until one succeeds
	for _, instance := range instances {
		err := rs.searchMovieOnInstance(ctx, instance, movieID)
		if err == nil {
			slog.Info("Movie search triggered successfully", "movieID", movieID, "instance", instance.Name)
			return nil
		}
		slog.Warn("Failed to trigger search on instance", "instance", instance.Name, "error", err)
	}

	return fmt.Errorf("failed to trigger search on any Radarr instance")
}

func (rs *radarrService) searchMovieOnInstance(ctx context.Context, instance repository.ArrService, movieID int) error {
	// Radarr command API to trigger movie search
	searchCommand := map[string]interface{}{
		"name":     "MoviesSearch",
		"movieIds": []int{movieID},
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

	resp, err := rs.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to contact Radarr: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("Radarr search command failed", 
			"movieID", movieID, 
			"status", resp.StatusCode, 
			"response", string(body),
			"instance", instance.Name)
		return fmt.Errorf("Radarr returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
