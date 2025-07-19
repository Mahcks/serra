package emby

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mahcks/serra/internal/global"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

type Service interface {
	getConfig() (string, string)
	GetLatestMedia(user *structures.User) ([]structures.EmbyMediaItem, error)
	GetAllLibraryItems() ([]structures.EmbyMediaItem, error)
	GetRecentlyAddedItems(maxAge string) ([]structures.EmbyMediaItem, error)
	GetEpisodesByTMDB(ctx context.Context, tmdbID int) ([]structures.EmbyMediaItem, error)
	GetEpisodesByTMDBAndSeason(ctx context.Context, tmdbID int, seasonNumber int) ([]structures.EmbyMediaItem, error)
	GetMovieByTMDBID(ctx context.Context, tmdbID int) (*structures.EmbyMediaItem, error)
	GetSeriesByTMDBID(ctx context.Context, tmdbID int) (*structures.EmbyMediaItem, error)
}

type embyService struct {
	gctx   global.Context
	client *http.Client
}

func New(gctx global.Context) Service {
	return &embyService{
		gctx:   gctx,
		client: utils.NewHTTPClient(),
	}
}

func (es *embyService) getConfig() (baseURL string, apiKey string) {
	cfg := es.gctx.Crate().Config.Get()
	baseURL = cfg.MediaServer.URL.String()
	apiKey = cfg.MediaServer.APIKey.String()

	return baseURL, apiKey
}

type baseItemDto struct {
	Name            string   `json:"Name"`
	OriginalTitle   string   `json:"OriginalTitle,omitempty"`
	ServerID        string   `json:"ServerId"`
	ID              string   `json:"Id"`
	ParentId        string   `json:"ParentId,omitempty"`
	SeriesId        string   `json:"SeriesId,omitempty"`
	SeasonId        string   `json:"SeasonId,omitempty"`
	SeasonNumber    int      `json:"ParentIndexNumber,omitempty"`
	EpisodeNumber   int      `json:"IndexNumber,omitempty"`
	RunTimeTicks    int64    `json:"RunTimeTicks"`
	IsFolder        bool     `json:"IsFolder"`
	Type            string   `json:"Type"`
	MediaType       string   `json:"MediaType,omitempty"`
	Status          string   `json:"Status,omitempty"`
	EndDate         string   `json:"EndDate,omitempty"`
	PremiereDate    string   `json:"PremiereDate,omitempty"`
	CommunityRating float64  `json:"CommunityRating,omitempty"`
	CriticRating    float64  `json:"CriticRating,omitempty"`
	OfficialRating  string   `json:"OfficialRating,omitempty"`
	Overview        string   `json:"Overview,omitempty"`
	Tagline         string   `json:"Tagline,omitempty"`
	Genres          []string `json:"Genres,omitempty"`
	Studios         []struct {
		Name string `json:"Name"`
		Id   string `json:"Id"`
	} `json:"Studios,omitempty"`
	People []struct {
		Name string `json:"Name"`
		Role string `json:"Role,omitempty"`
		Type string `json:"Type"`
	} `json:"People,omitempty"`
	UserData          userItemDataDto   `json:"UserData"`
	ImageTags         map[string]string `json:"ImageTags,omitempty"`
	BackdropImageTags []string          `json:"BackdropImageTags,omitempty"`
	AirDays           []string          `json:"AirDays,omitempty"`
	ProviderIds       map[string]string `json:"ProviderIds,omitempty"`
	ProductionYear    int               `json:"ProductionYear,omitempty"`
	Path              string            `json:"Path,omitempty"`
	Container         string            `json:"Container,omitempty"`
	Size              int64             `json:"Size,omitempty"`
	Bitrate           int               `json:"Bitrate,omitempty"`
	Width             int               `json:"Width,omitempty"`
	Height            int               `json:"Height,omitempty"`
	AspectRatio       string            `json:"AspectRatio,omitempty"`
	VideoType         string            `json:"VideoType,omitempty"`
	MediaStreams      []struct {
		Type      string `json:"Type"`
		Codec     string `json:"Codec,omitempty"`
		Language  string `json:"Language,omitempty"`
		Title     string `json:"Title,omitempty"`
		IsDefault bool   `json:"IsDefault,omitempty"`
		IsForced  bool   `json:"IsForced,omitempty"`
		Index     int    `json:"Index"`
	} `json:"MediaStreams,omitempty"`
	Tags             []string `json:"Tags,omitempty"`
	SortName         string   `json:"SortName,omitempty"`
	ForcedSortName   string   `json:"ForcedSortName,omitempty"`
	DateCreated      string   `json:"DateCreated,omitempty"`
	DateLastModified string   `json:"DateLastModified,omitempty"`
	IsHD             bool     `json:"IsHD,omitempty"`
	Locked           bool     `json:"LockedFields,omitempty"`
}

type userItemDataDto struct {
	PlayedPercentage      float64 `json:"PlayedPercentage,omitempty"`
	PlaybackPositionTicks int64   `json:"PlaybackPositionTicks"`
	PlayCount             int     `json:"PlayCount"`
	IsFavorite            bool    `json:"IsFavorite"`
	Played                bool    `json:"Played"`
	UnplayedItemCount     int     `json:"UnplayedItemCount,omitempty"`
}

func (es *embyService) GetLatestMedia(user *structures.User) ([]structures.EmbyMediaItem, error) {
	baseURL, _ := es.getConfig()

	// Include ProviderIds and other essential fields in the request
	fields := "ProviderIds,OriginalTitle,PremiereDate,CommunityRating,CriticRating,OfficialRating,Overview,Tagline,Genres,Studios,People,ProductionYear"
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/Users/%s/Items/Latest?Limit=15&Fields=%s", baseURL, user.ID, fields), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	// Set appropriate auth header based on server type
	serverType := es.gctx.Crate().Config.Get().MediaServer.Type
	if serverType == "jellyfin" {
		// Use the proper Authorization header for Jellyfin
		req.Header.Set("Authorization", fmt.Sprintf(`MediaBrowser Client="Serra", Device="Serra Dashboard", DeviceId="serra-dashboard", Version="1.0.0", Token="%s"`, user.AccessToken))
	} else {
		// Use X-Emby-Token for Emby
		req.Header.Set("X-Emby-Token", user.AccessToken)
	}

	resp, err := es.client.Do(req)
	if err != nil {
		return nil, apiErrors.ErrInternalServerError().SetDetail("Failed to fetch from Emby")
	}
	defer resp.Body.Close()

	var latestItems []baseItemDto
	if err := json.NewDecoder(resp.Body).Decode(&latestItems); err != nil {
		return nil, apiErrors.ErrInternalServerError().SetDetail("Failed to decode Emby response")
	}

	// Use the same conversion logic as other methods for consistency, but include all items (not just those with TMDB IDs)
	return es.convertItemsToEmbyMediaItemsIncludeAll(baseURL, latestItems), nil
}

func buildPrimaryPosterURL(baseURL string, item baseItemDto) string {
	primaryTag, hasPrimary := item.ImageTags["Primary"]
	if !hasPrimary || primaryTag == "" {
		return "" // or return a placeholder image URL
	}
	return fmt.Sprintf("%s/emby/Items/%s/Images/Primary?maxHeight=266&maxWidth=177&tag=%s", baseURL, item.ID, primaryTag)
}

// GetAllLibraryItems fetches all movies and TV shows from Emby/Jellyfin with TMDB IDs
func (es *embyService) GetAllLibraryItems() ([]structures.EmbyMediaItem, error) {
	baseURL, apiKey := es.getConfig()

	// Get all items of type Movie and Series with comprehensive metadata
	fields := "ProviderIds,Path,ProductionYear,OriginalTitle,PremiereDate,EndDate,CommunityRating,CriticRating,OfficialRating,Overview,Tagline,Genres,Studios,People,Container,Size,Bitrate,Width,Height,AspectRatio,MediaStreams,Tags,SortName,ForcedSortName,DateCreated,DateLastModified,IsHD"
	url := fmt.Sprintf("%s/Items?IncludeItemTypes=Movie,Series&Fields=%s&Recursive=true&api_key=%s", baseURL, fields, apiKey)

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

	return es.convertItemsToEmbyMediaItems(baseURL, response.Items), nil
}

// GetRecentlyAddedItems fetches recently added movies and TV shows from Emby/Jellyfin
// maxAge should be in format like "1 hour", "30 minutes", "1 day", etc.
func (es *embyService) GetRecentlyAddedItems(maxAge string) ([]structures.EmbyMediaItem, error) {
	baseURL, apiKey := es.getConfig()

	// Get recently added items with comprehensive metadata
	fields := "ProviderIds,Path,ProductionYear,OriginalTitle,PremiereDate,EndDate,CommunityRating,CriticRating,OfficialRating,Overview,Tagline,Genres,Studios,People,Container,Size,Bitrate,Width,Height,AspectRatio,MediaStreams,Tags,SortName,ForcedSortName,DateCreated,DateLastModified,IsHD"

	// Use DateCreated filter to get recently added items
	url := fmt.Sprintf("%s/Items?IncludeItemTypes=Movie,Series&Fields=%s&Recursive=true&MinDateCreated=%s&SortBy=DateCreated&SortOrder=Descending&api_key=%s",
		baseURL, fields, maxAge, apiKey)

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

	return es.convertItemsToEmbyMediaItems(baseURL, response.Items), nil
}

// convertItemsToEmbyMediaItemsIncludeAll converts baseItemDto slice to EmbyMediaItem slice including all items
// This is used for latest media which should show all items regardless of TMDB ID availability
func (es *embyService) convertItemsToEmbyMediaItemsIncludeAll(baseURL string, items []baseItemDto) []structures.EmbyMediaItem {
	var result []structures.EmbyMediaItem

	for _, item := range items {
		// Get provider IDs (may be empty for some items)
		tmdbID, _ := item.ProviderIds["Tmdb"]
		imdbID, _ := item.ProviderIds["Imdb"]
		tvdbID, _ := item.ProviderIds["Tvdb"]
		musicBrainzID, _ := item.ProviderIds["MusicBrainzAlbum"]

		// Determine media type
		mediaType := "movie"
		if item.Type == "Series" {
			mediaType = "tv"
		} else if item.Type == "Episode" {
			mediaType = "episode"
		}

		// Convert studios
		var studios []string
		for _, studio := range item.Studios {
			studios = append(studios, studio.Name)
		}

		// Convert people
		var people []structures.EmbyPerson
		for _, person := range item.People {
			people = append(people, structures.EmbyPerson{
				Name: person.Name,
				Role: person.Role,
				Type: person.Type,
			})
		}

		// Convert media streams to tracks
		var audioTracks, subtitleTracks []structures.EmbyMediaTrack
		for _, stream := range item.MediaStreams {
			track := structures.EmbyMediaTrack{
				Index:     stream.Index,
				Language:  stream.Language,
				Codec:     stream.Codec,
				Title:     stream.Title,
				IsDefault: stream.IsDefault,
				IsForced:  stream.IsForced,
			}

			switch stream.Type {
			case "Audio":
				audioTracks = append(audioTracks, track)
			case "Subtitle":
				subtitleTracks = append(subtitleTracks, track)
			}
		}

		// Calculate runtime in minutes
		runtimeMinutes := 0
		if item.RunTimeTicks > 0 {
			runtimeMinutes = int(item.RunTimeTicks / 600000000) // Convert ticks to minutes
		}

		// Determine quality flags
		isHD := item.IsHD || item.Height >= 720
		is4K := item.Height >= 2160

		// Build the enriched structure
		enrichedItem := structures.EmbyMediaItem{
			ID:              item.ID,
			Name:            item.Name,
			OriginalTitle:   item.OriginalTitle,
			Type:            mediaType,
			ParentID:        item.ParentId,
			SeriesID:        item.SeriesId,
			SeasonNumber:    item.SeasonNumber,
			EpisodeNumber:   item.EpisodeNumber,
			Year:            item.ProductionYear,
			PremiereDate:    item.PremiereDate,
			EndDate:         item.EndDate,
			CommunityRating: item.CommunityRating,
			CriticRating:    item.CriticRating,
			OfficialRating:  item.OfficialRating,
			Overview:        item.Overview,
			Tagline:         item.Tagline,
			Genres:          item.Genres,
			Studios:         studios,
			People:          people,
			TmdbID:          tmdbID,
			ImdbID:          imdbID,
			TvdbID:          tvdbID,
			MusicBrainzID:   musicBrainzID,
			ProviderIds:     item.ProviderIds, // Include the full provider IDs map
			Path:            item.Path,
			Container:       item.Container,
			SizeBytes:       item.Size,
			Bitrate:         item.Bitrate,
			Width:           item.Width,
			Height:          item.Height,
			AspectRatio:     item.AspectRatio,
			SubtitleTracks:  subtitleTracks,
			AudioTracks:     audioTracks,
			RuntimeTicks:    item.RunTimeTicks,
			RuntimeMinutes:  runtimeMinutes,
			IsFolder:        item.IsFolder,
			IsResumable:     item.UserData.PlaybackPositionTicks > 0,
			PlayCount:       item.UserData.PlayCount,
			DateCreated:     item.DateCreated,
			DateModified:    item.DateLastModified,
			IsHD:            isHD,
			Is4K:            is4K,
			Locked:          item.Locked,
			Tags:            item.Tags,
			SortName:        item.SortName,
			ForcedSortName:  item.ForcedSortName,

			// Legacy field for compatibility
			Poster: buildPrimaryPosterURL(baseURL, item),
		}

		result = append(result, enrichedItem)
	}

	return result
}

// convertItemsToEmbyMediaItems converts baseItemDto slice to EmbyMediaItem slice
// This extracts the common conversion logic used by both GetAllLibraryItems and GetRecentlyAddedItems
func (es *embyService) convertItemsToEmbyMediaItems(baseURL string, items []baseItemDto) []structures.EmbyMediaItem {
	var result []structures.EmbyMediaItem

	for _, item := range items {
		// Skip items without TMDB ID
		tmdbID, hasTmdbID := item.ProviderIds["Tmdb"]
		if !hasTmdbID || tmdbID == "" {
			continue
		}

		// Determine media type
		mediaType := "movie"
		if item.Type == "Series" {
			mediaType = "tv"
		}

		// Convert studios
		var studios []string
		for _, studio := range item.Studios {
			studios = append(studios, studio.Name)
		}

		// Convert people
		var people []structures.EmbyPerson
		for _, person := range item.People {
			people = append(people, structures.EmbyPerson{
				Name: person.Name,
				Role: person.Role,
				Type: person.Type,
			})
		}

		// Convert media streams to tracks
		var audioTracks, subtitleTracks []structures.EmbyMediaTrack
		for _, stream := range item.MediaStreams {
			track := structures.EmbyMediaTrack{
				Index:     stream.Index,
				Language:  stream.Language,
				Codec:     stream.Codec,
				Title:     stream.Title,
				IsDefault: stream.IsDefault,
				IsForced:  stream.IsForced,
			}

			switch stream.Type {
			case "Audio":
				audioTracks = append(audioTracks, track)
			case "Subtitle":
				subtitleTracks = append(subtitleTracks, track)
			}
		}

		// Get other provider IDs
		imdbID, _ := item.ProviderIds["Imdb"]
		tvdbID, _ := item.ProviderIds["Tvdb"]
		musicBrainzID, _ := item.ProviderIds["MusicBrainzAlbum"]

		// Calculate runtime in minutes
		runtimeMinutes := 0
		if item.RunTimeTicks > 0 {
			runtimeMinutes = int(item.RunTimeTicks / 600000000) // Convert ticks to minutes
		}

		// Determine quality flags
		isHD := item.IsHD || item.Height >= 720
		is4K := item.Height >= 2160

		// Build the enriched structure
		enrichedItem := structures.EmbyMediaItem{
			ID:              item.ID,
			Name:            item.Name,
			OriginalTitle:   item.OriginalTitle,
			Type:            mediaType,
			ParentID:        item.ParentId,
			SeriesID:        item.SeriesId,
			SeasonNumber:    item.SeasonNumber,
			EpisodeNumber:   item.EpisodeNumber,
			Year:            item.ProductionYear,
			PremiereDate:    item.PremiereDate,
			EndDate:         item.EndDate,
			CommunityRating: item.CommunityRating,
			CriticRating:    item.CriticRating,
			OfficialRating:  item.OfficialRating,
			Overview:        item.Overview,
			Tagline:         item.Tagline,
			Genres:          item.Genres,
			Studios:         studios,
			People:          people,
			TmdbID:          tmdbID,
			ImdbID:          imdbID,
			TvdbID:          tvdbID,
			MusicBrainzID:   musicBrainzID,
			ProviderIds:     item.ProviderIds, // Include the full provider IDs map
			Path:            item.Path,
			Container:       item.Container,
			SizeBytes:       item.Size,
			Bitrate:         item.Bitrate,
			Width:           item.Width,
			Height:          item.Height,
			AspectRatio:     item.AspectRatio,
			SubtitleTracks:  subtitleTracks,
			AudioTracks:     audioTracks,
			RuntimeTicks:    item.RunTimeTicks,
			RuntimeMinutes:  runtimeMinutes,
			IsFolder:        item.IsFolder,
			IsResumable:     item.UserData.PlaybackPositionTicks > 0,
			PlayCount:       item.UserData.PlayCount,
			DateCreated:     item.DateCreated,
			DateModified:    item.DateLastModified,
			IsHD:            isHD,
			Is4K:            is4K,
			Locked:          item.Locked,
			Tags:            item.Tags,
			SortName:        item.SortName,
			ForcedSortName:  item.ForcedSortName,

			// Legacy field for compatibility
			Poster: buildPrimaryPosterURL(baseURL, item),
		}

		result = append(result, enrichedItem)
	}

	return result
}
