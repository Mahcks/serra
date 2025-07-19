package emby

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mahcks/serra/pkg/structures"
)

// GetEpisodesByTMDB fetches all episodes for a TV show by TMDB ID
func (es *embyService) GetEpisodesByTMDB(ctx context.Context, tmdbID int) ([]structures.EmbyMediaItem, error) {
	baseURL, apiKey := es.getConfig()

	log.Printf("🔍 [Emby] Searching for series with TMDB ID %d", tmdbID)

	// Get all episodes for the series with the given TMDB ID
	fields := "ProviderIds,Path,ProductionYear,OriginalTitle,PremiereDate,EndDate,CommunityRating,CriticRating,OfficialRating,Overview,Tagline,Genres,Studios,People,Container,Size,Bitrate,Width,Height,AspectRatio,MediaStreams,Tags,SortName,ForcedSortName,DateCreated,DateLastModified,IsHD,ParentIndexNumber,IndexNumber,SeriesId,SeasonId"
	
	// First, find the series with this TMDB ID
	seriesURL := fmt.Sprintf("%s/Items?IncludeItemTypes=Series&Fields=%s&Recursive=true&api_key=%s", baseURL, fields, apiKey)
	log.Printf("🔍 [Emby] Series search URL: %s", seriesURL)
	
	seriesReq, err := http.NewRequestWithContext(ctx, "GET", seriesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create series request: %w", err)
	}

	seriesResp, err := es.client.Do(seriesReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch series from Emby: %w", err)
	}
	defer seriesResp.Body.Close()

	var seriesResponse struct {
		Items []baseItemDto `json:"Items"`
	}

	if err := json.NewDecoder(seriesResp.Body).Decode(&seriesResponse); err != nil {
		return nil, fmt.Errorf("failed to decode series response: %w", err)
	}

	log.Printf("📺 [Emby] Found %d series total", len(seriesResponse.Items))

	// Log first few series for debugging
	for i, series := range seriesResponse.Items {
		if i < 5 {
			tmdbID := "none"
			if series.ProviderIds != nil {
				if id, exists := series.ProviderIds["Tmdb"]; exists {
					tmdbID = id
				}
			}
			log.Printf("   Series %d: %s (TMDB: %s)", i+1, series.Name, tmdbID)
		}
	}

	// Find the series with matching TMDB ID
	var seriesID string
	for _, series := range seriesResponse.Items {
		if tmdbIDStr, exists := series.ProviderIds["Tmdb"]; exists && tmdbIDStr == fmt.Sprintf("%d", tmdbID) {
			seriesID = series.ID
			log.Printf("✅ [Emby] Found matching series: %s (ID: %s, TMDB: %s)", series.Name, series.ID, tmdbIDStr)
			break
		}
	}

	if seriesID == "" {
		log.Printf("❌ [Emby] No series found with TMDB ID %d", tmdbID)
		return []structures.EmbyMediaItem{}, nil // No series found with this TMDB ID
	}

	// Now get all episodes for this series
	episodesURL := fmt.Sprintf("%s/Shows/%s/Episodes?Fields=%s&api_key=%s", baseURL, seriesID, fields, apiKey)
	log.Printf("🔍 [Emby] Episodes search URL: %s", episodesURL)
	
	episodesReq, err := http.NewRequestWithContext(ctx, "GET", episodesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create episodes request: %w", err)
	}

	episodesResp, err := es.client.Do(episodesReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch episodes from Emby: %w", err)
	}
	defer episodesResp.Body.Close()

	var episodesResponse struct {
		Items []baseItemDto `json:"Items"`
	}

	if err := json.NewDecoder(episodesResp.Body).Decode(&episodesResponse); err != nil {
		return nil, fmt.Errorf("failed to decode episodes response: %w", err)
	}

	log.Printf("📺 [Emby] Found %d episodes for series ID %s", len(episodesResponse.Items), seriesID)
	return es.convertItemsToEmbyMediaItemsIncludeAll(baseURL, episodesResponse.Items), nil
}

// GetEpisodesByTMDBAndSeason fetches episodes for a specific season of a TV show by TMDB ID
func (es *embyService) GetEpisodesByTMDBAndSeason(ctx context.Context, tmdbID int, seasonNumber int) ([]structures.EmbyMediaItem, error) {
	// Get all episodes for the series
	allEpisodes, err := es.GetEpisodesByTMDB(ctx, tmdbID)
	if err != nil {
		return nil, err
	}

	// Filter episodes by season number
	var seasonEpisodes []structures.EmbyMediaItem
	for _, episode := range allEpisodes {
		if episode.SeasonNumber == seasonNumber {
			seasonEpisodes = append(seasonEpisodes, episode)
		}
	}

	return seasonEpisodes, nil
}