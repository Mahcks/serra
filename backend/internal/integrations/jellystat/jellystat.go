package jellystat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

var jellystatNotEnabled = fmt.Errorf("jellystat not enabled: returning empty array")

type Service interface {
	GetLibraryOverview() ([]structures.JellystatLibrary, error)
	GetUserActivity() ([]structures.JellystatUserActivity, error)
}

type jellystatService struct {
	gctx   global.Context
	client *http.Client
}

func New(gctx global.Context) Service {
	return &jellystatService{
		gctx: gctx,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type jellyStatLibraryItem struct {
	ID             string `json:"Id"`
	Name           string `json:"Name"`
	CollectionType string `json:"CollectionType"`
	LibraryCount   int    `json:"Library_Count"`
	SeasonCount    int    `json:"Season_Count"`
	EpisodeCount   int    `json:"Episode_Count"`
}

func (j *jellystatService) GetLibraryOverview() ([]structures.JellystatLibrary, error) {
	cfg := j.gctx.Crate().Config.Get()
	baseURL := cfg.Jellystat.URL.String()
	apiKey := cfg.Jellystat.APIKey.String()

	if baseURL == "" {
		return utils.EmptyResult[structures.JellystatLibrary](jellystatNotEnabled.Error())
	}

	req, err := http.NewRequest("GET", baseURL+"/stats/getLibraryOverview", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Token", apiKey)

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch Jellystat library overview")
	}

	var libraries []jellyStatLibraryItem
	if err := json.NewDecoder(resp.Body).Decode(&libraries); err != nil {
		return nil, fmt.Errorf("failed to decode Jellystat response: %w", err)
	}

	var result []structures.JellystatLibrary
	for _, lib := range libraries {
		result = append(result, structures.JellystatLibrary{
			ID:             lib.ID,
			Name:           lib.Name,
			CollectionType: lib.CollectionType,
			LibraryCount:   lib.LibraryCount,
			SeasonCount:    lib.SeasonCount,
			EpisodeCount:   lib.EpisodeCount,
		})
	}

	return result, nil
}

type jellystatUserActivity struct {
	UserID         string `json:"UserId"`
	UserName       string `json:"UserName"`
	TotalPlays     int    `json:"TotalPlays"`
	TotalWatchTime int    `json:"TotalWatchTime"`
}

func (j *jellystatService) GetUserActivity() ([]structures.JellystatUserActivity, error) {
	cfg := j.gctx.Crate().Config.Get()
	baseURL := cfg.Jellystat.URL.String()
	apiKey := cfg.Jellystat.APIKey.String()

	if baseURL == "" {
		return utils.EmptyResult[structures.JellystatUserActivity](jellystatNotEnabled.Error())
	}

	req, err := http.NewRequest("GET", baseURL+"/stats/getAllUserActivity", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Token", apiKey)

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch Jellystat library overview")
	}

	var users []jellystatUserActivity
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode Jellystat response: %w", err)
	}

	var result []structures.JellystatUserActivity
	for _, user := range users {
		result = append(result, structures.JellystatUserActivity{
			UserID:         user.UserID,
			UserName:       user.UserName,
			TotalPlays:     user.TotalPlays,
			TotalWatchTime: user.TotalWatchTime,
		})
	}

	return result, nil
}
