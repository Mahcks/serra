package sonarr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/pkg/structures"
)

type Service interface {
	GetUpcomingItems(ctx context.Context) ([]structures.CalendarItem, error)
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
				ID    int    `json:"id"`
				Title string `json:"title"`
			}
			if err := json.NewDecoder(seriesResp.Body).Decode(&seriesList); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("instance %s: failed to decode Sonarr series response: %w", inst.Name, err))
				mu.Unlock()
				return
			}

			// Build a quick lookup map
			seriesMap := make(map[int]string)
			for _, s := range seriesList {
				seriesMap[s.ID] = s.Title
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

					fullTitle := fmt.Sprintf("%s S%02dE%02d - %s",
						seriesTitle, r.SeasonNumber, r.EpisodeNumber, r.Title)

					items = append(items, structures.CalendarItem{
						Title:       fullTitle,
						Source:      structures.ProviderSonarr,
						ReleaseDate: r.AirDateUtc,
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
