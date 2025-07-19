package emby

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mahcks/serra/pkg/structures"
)

// GetMovieByTMDBID fetches a specific movie by TMDB ID from Emby/Jellyfin
func (es *embyService) GetMovieByTMDBID(ctx context.Context, tmdbID int) (*structures.EmbyMediaItem, error) {
	baseURL, apiKey := es.getConfig()

	// Search for movie by TMDB ID
	fields := "ProviderIds,Path,ProductionYear,OriginalTitle,PremiereDate,CommunityRating,CriticRating,OfficialRating,Overview,Tagline,Genres,Studios,People,Container,Size,Bitrate,Width,Height,AspectRatio,MediaStreams,Tags,SortName,ForcedSortName,DateCreated,DateLastModified,IsHD"
	url := fmt.Sprintf("%s/Items?IncludeItemTypes=Movie&Fields=%s&Recursive=true&AnyProviderIdEquals=tmdb.%d&api_key=%s", baseURL, fields, tmdbID, apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := es.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from Emby: %w", err)
	}
	defer resp.Body.Close()

	var response struct {
		Items            []baseItemDto `json:"Items"`
		TotalRecordCount int           `json:"TotalRecordCount"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode Emby response: %w", err)
	}

	if len(response.Items) == 0 {
		return nil, nil // Movie not found
	}

	// Convert first match to EmbyMediaItem
	items := es.convertItemsToEmbyMediaItems(baseURL, response.Items)
	if len(items) == 0 {
		return nil, nil // No valid items found
	}

	return &items[0], nil
}

// GetSeriesByTMDBID fetches a specific TV series by TMDB ID from Emby/Jellyfin
func (es *embyService) GetSeriesByTMDBID(ctx context.Context, tmdbID int) (*structures.EmbyMediaItem, error) {
	baseURL, apiKey := es.getConfig()

	// Search for series by TMDB ID
	fields := "ProviderIds,Path,ProductionYear,OriginalTitle,PremiereDate,EndDate,CommunityRating,CriticRating,OfficialRating,Overview,Tagline,Genres,Studios,People,Container,Size,Bitrate,Width,Height,AspectRatio,MediaStreams,Tags,SortName,ForcedSortName,DateCreated,DateLastModified,IsHD"
	url := fmt.Sprintf("%s/Items?IncludeItemTypes=Series&Fields=%s&Recursive=true&AnyProviderIdEquals=tmdb.%d&api_key=%s", baseURL, fields, tmdbID, apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := es.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from Emby: %w", err)
	}
	defer resp.Body.Close()

	var response struct {
		Items            []baseItemDto `json:"Items"`
		TotalRecordCount int           `json:"TotalRecordCount"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode Emby response: %w", err)
	}

	if len(response.Items) == 0 {
		return nil, nil // Series not found
	}

	// Convert first match to EmbyMediaItem
	items := es.convertItemsToEmbyMediaItems(baseURL, response.Items)
	if len(items) == 0 {
		return nil, nil // No valid items found
	}

	return &items[0], nil
}