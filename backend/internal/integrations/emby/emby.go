package emby

import (
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
	Name              string            `json:"Name"`
	ServerID          string            `json:"ServerId"`
	ID                string            `json:"Id"`
	RunTimeTicks      int64             `json:"RunTimeTicks"`
	IsFolder          bool              `json:"IsFolder"`
	Type              string            `json:"Type"`
	MediaType         string            `json:"MediaType,omitempty"`
	Status            string            `json:"Status,omitempty"`
	EndDate           string            `json:"EndDate,omitempty"`
	UserData          userItemDataDto   `json:"UserData"`
	ImageTags         map[string]string `json:"ImageTags,omitempty"`
	BackdropImageTags []string          `json:"BackdropImageTags,omitempty"`
	AirDays           []string          `json:"AirDays,omitempty"`
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

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/Users/%s/Items/Latest?Limit=15", baseURL, user.ID), nil)
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

	var result []structures.EmbyMediaItem
	for _, item := range latestItems {
		poster := buildPrimaryPosterURL(baseURL, item)

		result = append(result, structures.EmbyMediaItem{
			ID:     item.ID,
			Name:   item.Name,
			Type:   item.Type,
			Poster: poster,
		})
	}

	return result, nil
}

func buildPrimaryPosterURL(baseURL string, item baseItemDto) string {
	primaryTag, hasPrimary := item.ImageTags["Primary"]
	if !hasPrimary || primaryTag == "" {
		return "" // or return a placeholder image URL
	}
	return fmt.Sprintf("%s/emby/Items/%s/Images/Primary?maxHeight=266&maxWidth=177&tag=%s", baseURL, item.ID, primaryTag)
}
