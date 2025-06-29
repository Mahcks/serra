package radarr

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
	GetCalendarItems(ctx context.Context) ([]structures.CalendarItem, error)
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
