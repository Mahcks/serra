package jellystat

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

var jellystatNotEnabled = fmt.Errorf("jellystat not enabled: returning empty array")

type Service interface {
	GetLibraryOverview() ([]structures.JellystatLibrary, error)
	GetUserActivity() ([]structures.JellystatUserActivity, error)
	GetPopularContent(limit int) ([]structures.JellystatPopularContent, error)
	GetMostViewedContent(limit int) ([]structures.JellystatPopularContent, error)
	GetMostActiveUsers(days int) ([]structures.JellystatActiveUser, error)
	GetPlaybackMethodStats(days int) ([]structures.JellystatPlaybackMethod, error)
	GetRecentlyWatched(limit int) ([]structures.JellystatRecentlyWatched, error)
	GetUserWatchHistory(userID string, limit int) ([]structures.JellystatWatchHistory, error)
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

// getJellystatConfig reads Jellystat settings from the database and constructs the configuration
func (j *jellystatService) getJellystatConfig(ctx context.Context) (enabled bool, baseURL, apiKey string, err error) {
	// Check if Jellystat is enabled
	enabledStr, err := j.gctx.Crate().Sqlite.Query().GetSetting(ctx, structures.SettingJellystatEnabled.String())
	if err != nil || enabledStr != "true" {
		return false, "", "", nil // Not enabled, return empty config
	}

	// Get API key
	apiKey, err = j.gctx.Crate().Sqlite.Query().GetSetting(ctx, structures.SettingJellystatAPIKey.String())
	if err != nil {
		return false, "", "", fmt.Errorf("failed to get Jellystat API key: %w", err)
	}

	// Check if full URL is provided (overrides host/port/ssl)
	fullURL, err := j.gctx.Crate().Sqlite.Query().GetSetting(ctx, structures.SettingJellystatURL.String())
	if err == nil && fullURL != "" {
		return true, fullURL, apiKey, nil
	}

	// Construct URL from host, port, and SSL settings
	host, err := j.gctx.Crate().Sqlite.Query().GetSetting(ctx, structures.SettingJellystatHost.String())
	if err != nil || host == "" {
		return false, "", "", fmt.Errorf("Jellystat host not configured")
	}

	port, err := j.gctx.Crate().Sqlite.Query().GetSetting(ctx, structures.SettingJellystatPort.String())
	if err != nil || port == "" {
		port = "3000" // Default port
	}

	useSSLStr, err := j.gctx.Crate().Sqlite.Query().GetSetting(ctx, structures.SettingJellystatUseSSL.String())
	if err != nil {
		useSSLStr = "false" // Default to HTTP
	}

	protocol := "http"
	if useSSLStr == "true" {
		protocol = "https"
	}

	baseURL = fmt.Sprintf("%s://%s:%s", protocol, host, port)
	return true, baseURL, apiKey, nil
}

type jellyStatLibraryItem struct {
	ID             string `json:"Id"`
	Name           string `json:"Name"`
	CollectionType string `json:"CollectionType"`
	LibraryCount   int    `json:"Library_Count"`
	SeasonCount    int    `json:"Season_Count"`
	EpisodeCount   int    `json:"Episode_Count"`
}

type jellystatPopularItem struct {
	ID                 string `json:"Id"`
	Name               string `json:"Name"`
	UniqueViewers      int    `json:"unique_viewers"`
	LatestActivityDate string `json:"latest_activity_date"`
	PrimaryImageHash   string `json:"PrimaryImageHash"`
	Archived           bool   `json:"archived"`
}

type jellystatViewedItem struct {
	ID                    string `json:"Id"`
	Name                  string `json:"Name"`
	Plays                 int    `json:"Plays"`
	TotalPlaybackDuration int    `json:"total_playback_duration"`
	PrimaryImageHash      string `json:"PrimaryImageHash"`
	Archived              bool   `json:"archived"`
}

type jellystatActiveUserItem struct {
	UserID string `json:"UserId"`
	Name   string `json:"Name"`
	Plays  int    `json:"Plays"`
}

type jellystatPlaybackMethodItem struct {
	Name  string `json:"Name"`
	Count int    `json:"Count"`
}

func (j *jellystatService) GetLibraryOverview() ([]structures.JellystatLibrary, error) {
	ctx := context.Background()
	enabled, baseURL, apiKey, err := j.getJellystatConfig(ctx)
	if err != nil {
		return nil, err
	}

	if !enabled || baseURL == "" {
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
		return nil, fmt.Errorf("Jellystat library overview API returned status %d", resp.StatusCode)
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
	ctx := context.Background()
	enabled, baseURL, apiKey, err := j.getJellystatConfig(ctx)
	if err != nil {
		return nil, err
	}

	if !enabled || baseURL == "" {
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
		return nil, fmt.Errorf("Jellystat user activity API returned status %d", resp.StatusCode)
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

// GetPopularContent returns the most popular content by play count
func (j *jellystatService) GetPopularContent(limit int) ([]structures.JellystatPopularContent, error) {
	ctx := context.Background()
	enabled, baseURL, apiKey, err := j.getJellystatConfig(ctx)
	if err != nil {
		slog.Error("Failed to get Jellystat configuration", "method", "GetPopularContent", "error", err)
		return nil, err
	}

	slog.Debug("Jellystat configuration loaded", "method", "GetPopularContent", "enabled", enabled, "baseURL", baseURL, "hasApiKey", apiKey != "")

	if !enabled || baseURL == "" {
		return utils.EmptyResult[structures.JellystatPopularContent](jellystatNotEnabled.Error())
	}

	// Get popular movies and TV shows separately
	var allItems []structures.JellystatPopularContent

	// Get popular movies
	movieBody := `{"days":30,"type":"Movie"}`
	movieReq, err := http.NewRequest("POST", baseURL+"/stats/getMostPopularByType", strings.NewReader(movieBody))
	if err != nil {
		return allItems, nil // Return empty if request creation fails
	}
	movieReq.Header.Set("X-API-Token", apiKey)
	movieReq.Header.Set("Content-Type", "application/json")
	movieReq.Header.Set("Accept", "application/json, text/plain, */*")

	slog.Debug("Fetching popular movies from Jellystat", "url", baseURL+"/stats/getMostPopularByType", "days", 30)

	movieResp, err := j.client.Do(movieReq)
	if err != nil {
		slog.Error("Failed to fetch popular movies from Jellystat", "error", err)
		return allItems, nil // Return empty if movies fail
	}
	defer movieResp.Body.Close()

	slog.Debug("Received response from Jellystat movies API", "status", movieResp.StatusCode)

	if movieResp.StatusCode != http.StatusOK {
		// Read response body for error details
		bodyBytes := make([]byte, 512) // Read first 512 bytes
		n, _ := movieResp.Body.Read(bodyBytes)
		slog.Error("Jellystat movies API returned error", "status", movieResp.StatusCode, "response", string(bodyBytes[:n]))
		return allItems, nil
	}

	if movieResp.StatusCode == http.StatusOK {
		var movieItems []jellystatPopularItem
		if err := json.NewDecoder(movieResp.Body).Decode(&movieItems); err == nil {
			for _, item := range movieItems {
				allItems = append(allItems, structures.JellystatPopularContent{
					ItemID:      item.ID,
					ItemName:    item.Name,
					ItemType:    "Movie",
					TotalPlays:  item.UniqueViewers, // Map unique viewers to total plays
					LibraryName: "Movies",
				})
			}
		}
	}

	// Get popular TV shows
	seriesBody := `{"days":30,"type":"Series"}`
	seriesReq, err := http.NewRequest("POST", baseURL+"/stats/getMostPopularByType", strings.NewReader(seriesBody))
	if err != nil {
		return allItems, nil // Return movies if series fails
	}
	seriesReq.Header.Set("X-API-Token", apiKey)
	seriesReq.Header.Set("Content-Type", "application/json")
	seriesReq.Header.Set("Accept", "application/json, text/plain, */*")

	seriesResp, err := j.client.Do(seriesReq)
	if err != nil {
		return allItems, nil // Return movies if series fails
	}
	defer seriesResp.Body.Close()

	if seriesResp.StatusCode == http.StatusOK {
		var seriesItems []jellystatPopularItem
		if err := json.NewDecoder(seriesResp.Body).Decode(&seriesItems); err != nil {
			return allItems, nil // Return movies if series decode fails
		}
		for _, item := range seriesItems {
			allItems = append(allItems, structures.JellystatPopularContent{
				ItemID:      item.ID,
				ItemName:    item.Name,
				ItemType:    "Series",
				TotalPlays:  item.UniqueViewers, // Map unique viewers to total plays
				LibraryName: "TV Shows",
			})
		}
	}

	// Limit results if requested
	if limit > 0 && len(allItems) > limit {
		allItems = allItems[:limit]
	}

	return allItems, nil
}

// GetMostViewedContent returns the most viewed content by view count
func (j *jellystatService) GetMostViewedContent(limit int) ([]structures.JellystatPopularContent, error) {
	ctx := context.Background()
	enabled, baseURL, apiKey, err := j.getJellystatConfig(ctx)
	if err != nil {
		slog.Error("Failed to get Jellystat configuration", "method", "GetMostViewedContent", "error", err)
		return nil, err
	}

	slog.Debug("Jellystat configuration loaded", "method", "GetMostViewedContent", "enabled", enabled, "baseURL", baseURL, "hasApiKey", apiKey != "")

	if !enabled || baseURL == "" {
		return utils.EmptyResult[structures.JellystatPopularContent](jellystatNotEnabled.Error())
	}

	// Get most viewed movies and TV shows separately
	var allItems []structures.JellystatPopularContent

	// Get most viewed movies
	movieBody := `{"days":30,"type":"Movie"}`
	movieReq, err := http.NewRequest("POST", baseURL+"/stats/getMostViewedByType", strings.NewReader(movieBody))
	if err != nil {
		return allItems, nil // Return empty if request creation fails
	}
	movieReq.Header.Set("X-API-Token", apiKey)
	movieReq.Header.Set("Content-Type", "application/json")
	movieReq.Header.Set("Accept", "application/json, text/plain, */*")

	slog.Debug("Fetching most viewed movies from Jellystat", "url", baseURL+"/stats/getMostViewedByType", "days", 30)

	movieResp, err := j.client.Do(movieReq)
	if err != nil {
		slog.Error("Failed to fetch most viewed movies from Jellystat", "error", err)
		return allItems, nil // Return empty if movies fail
	}
	defer movieResp.Body.Close()

	slog.Debug("Received response from Jellystat most viewed movies API", "status", movieResp.StatusCode)

	if movieResp.StatusCode != http.StatusOK {
		// Read response body for error details
		bodyBytes := make([]byte, 512) // Read first 512 bytes
		n, _ := movieResp.Body.Read(bodyBytes)
		slog.Error("Jellystat most viewed movies API returned error", "status", movieResp.StatusCode, "response", string(bodyBytes[:n]))
		return allItems, nil
	}

	if movieResp.StatusCode == http.StatusOK {
		var movieItems []jellystatViewedItem
		if err := json.NewDecoder(movieResp.Body).Decode(&movieItems); err == nil {
			for _, item := range movieItems {
				allItems = append(allItems, structures.JellystatPopularContent{
					ItemID:      item.ID,
					ItemName:    item.Name,
					ItemType:    "Movie",
					TotalPlays:  item.Plays, // Use Plays field from viewed API
					LibraryName: "Movies",
				})
			}
		}
	}

	// Get most viewed TV shows
	seriesBody := `{"days":30,"type":"Series"}`
	seriesReq, err := http.NewRequest("POST", baseURL+"/stats/getMostViewedByType", strings.NewReader(seriesBody))
	if err != nil {
		return allItems, nil // Return movies if series fails
	}
	seriesReq.Header.Set("X-API-Token", apiKey)
	seriesReq.Header.Set("Content-Type", "application/json")
	seriesReq.Header.Set("Accept", "application/json, text/plain, */*")

	seriesResp, err := j.client.Do(seriesReq)
	if err != nil {
		return allItems, nil // Return movies if series fails
	}
	defer seriesResp.Body.Close()

	if seriesResp.StatusCode == http.StatusOK {
		var seriesItems []jellystatViewedItem
		if err := json.NewDecoder(seriesResp.Body).Decode(&seriesItems); err == nil {
			for _, item := range seriesItems {
				allItems = append(allItems, structures.JellystatPopularContent{
					ItemID:      item.ID,
					ItemName:    item.Name,
					ItemType:    "Series",
					TotalPlays:  item.Plays, // Use Plays field from viewed API
					LibraryName: "TV Shows",
				})
			}
		}
	}

	// Sort by total plays (descending order)
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].TotalPlays > allItems[j].TotalPlays
	})

	// Limit results if requested
	if limit > 0 && len(allItems) > limit {
		allItems = allItems[:limit]
	}

	return allItems, nil
}

// GetMostActiveUsers returns the most active users by play count
func (j *jellystatService) GetMostActiveUsers(days int) ([]structures.JellystatActiveUser, error) {
	ctx := context.Background()
	enabled, baseURL, apiKey, err := j.getJellystatConfig(ctx)
	if err != nil {
		slog.Error("Failed to get Jellystat configuration", "method", "GetMostActiveUsers", "error", err)
		return nil, err
	}

	if !enabled || baseURL == "" {
		return utils.EmptyResult[structures.JellystatActiveUser](jellystatNotEnabled.Error())
	}

	requestBody := fmt.Sprintf(`{"days":%d}`, days)
	req, err := http.NewRequest("POST", baseURL+"/stats/getMostActiveUsers", strings.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-API-Token", apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active users: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var apiUsers []jellystatActiveUserItem
	if err := json.NewDecoder(resp.Body).Decode(&apiUsers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var users []structures.JellystatActiveUser
	for _, user := range apiUsers {
		users = append(users, structures.JellystatActiveUser{
			UserID:   user.UserID,
			UserName: user.Name,
			Plays:    user.Plays,
		})
	}

	return users, nil
}

// GetPlaybackMethodStats returns playback method statistics
func (j *jellystatService) GetPlaybackMethodStats(days int) ([]structures.JellystatPlaybackMethod, error) {
	ctx := context.Background()
	enabled, baseURL, apiKey, err := j.getJellystatConfig(ctx)
	if err != nil {
		slog.Error("Failed to get Jellystat configuration", "method", "GetPlaybackMethodStats", "error", err)
		return nil, err
	}

	if !enabled || baseURL == "" {
		return utils.EmptyResult[structures.JellystatPlaybackMethod](jellystatNotEnabled.Error())
	}

	requestBody := fmt.Sprintf(`{"days":%d}`, days)
	req, err := http.NewRequest("POST", baseURL+"/stats/getPlaybackMethodStats", strings.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-API-Token", apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch playback stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var apiMethods []jellystatPlaybackMethodItem
	if err := json.NewDecoder(resp.Body).Decode(&apiMethods); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var methods []structures.JellystatPlaybackMethod
	for _, method := range apiMethods {
		methods = append(methods, structures.JellystatPlaybackMethod{
			Name:  method.Name,
			Count: method.Count,
		})
	}

	return methods, nil
}

// GetRecentlyWatched returns recently watched content across all users
func (j *jellystatService) GetRecentlyWatched(limit int) ([]structures.JellystatRecentlyWatched, error) {
	// TODO: Find the correct Jellystat API endpoint for recently watched content
	// The current Jellystat API doesn't seem to have a direct recently watched endpoint
	// For now, return empty results to avoid errors
	return []structures.JellystatRecentlyWatched{}, nil
}

// GetUserWatchHistory returns watch history for a specific user
func (j *jellystatService) GetUserWatchHistory(userID string, limit int) ([]structures.JellystatWatchHistory, error) {
	ctx := context.Background()
	enabled, baseURL, apiKey, err := j.getJellystatConfig(ctx)
	if err != nil {
		return nil, err
	}

	if !enabled || baseURL == "" {
		return utils.EmptyResult[structures.JellystatWatchHistory](jellystatNotEnabled.Error())
	}

	endpoint := fmt.Sprintf("/stats/getUserWatchHistory?userId=%s&limit=%d", userID, limit)
	req, err := http.NewRequest("GET", baseURL+endpoint, nil)
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
		return nil, fmt.Errorf("Jellystat user watch history API returned status %d", resp.StatusCode)
	}

	var items []structures.JellystatWatchHistory
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("failed to decode user watch history response: %w", err)
	}

	return items, nil
}
