package sonarr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

type getProfilesRequest struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
}

type qualityProfile struct {
	ID                    int                  `json:"id"`
	Name                  string               `json:"name"`
	UpgradeAllowed        bool                 `json:"upgradeAllowed"`
	Cutoff                int                  `json:"cutoff"`
	Items                 []qualityProfileItem `json:"items"`
	MinFormatScore        int                  `json:"minFormatScore"`
	CutoffFormatScore     int                  `json:"cutoffFormatScore"`
	MinUpgradeFormatScore int                  `json:"minUpgradeFormatScore"`
	FormatItems           []formatItem         `json:"formatItems"`
	Language              language             `json:"language"`
}

type qualityProfileItem struct {
	Quality quality              `json:"quality"`
	Items   []qualityProfileItem `json:"items"`
	Allowed bool                 `json:"allowed"`
	Name    string               `json:"name,omitempty"` // Only present for grouped items
	ID      int                  `json:"id,omitempty"`   // Only present for grouped items
}

type quality struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Source     string `json:"source"`
	Resolution int    `json:"resolution"`
	Modifier   string `json:"modifier"`
}

type formatItem struct {
	Format int    `json:"format"`
	Name   string `json:"name"`
	Score  int    `json:"score"`
}

type language struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (rg *RouteGroup) GetProfiles(ctx *respond.Ctx) error {
	var req getProfilesRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("failed to parse request body")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	url := fmt.Sprintf("%s/api/v3/qualityprofile", req.BaseURL)
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to create request")
	}
	httpReq.Header.Set("X-Api-Key", req.APIKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		return apiErrors.ErrBadGateway().SetDetail("failed to contact Sonarr server")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return apiErrors.ErrInternalServerError().SetDetail("Sonarr server returned an error: " + resp.Status)
	}

	var profiles []qualityProfile
	err = json.NewDecoder(resp.Body).Decode(&profiles)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to decode Sonarr response")
	}

	var result []structures.SonarrQualityProfile
	for _, profile := range profiles {
		sonarrProfile := structures.SonarrQualityProfile{
			ID:                    profile.ID,
			Name:                  profile.Name,
			UpgradeAllowed:        profile.UpgradeAllowed,
			Cutoff:                profile.Cutoff,
			MinFormatScore:        profile.MinFormatScore,
			CutoffFormatScore:     profile.CutoffFormatScore,
			MinUpgradeFormatScore: profile.MinUpgradeFormatScore,
			Items:                 make([]structures.SonarrQualityProfileItem, 0, len(profile.Items)),
			FormatItems:           make([]structures.SonarrFormatItem, 0, len(profile.FormatItems)),
			Language: structures.SonarrLanguage{
				ID:   profile.Language.ID,
				Name: profile.Language.Name,
			},
		}
		for _, item := range profile.Items {
			sonarrItem := structures.SonarrQualityProfileItem{
				Quality: structures.SonarrQuality{
					ID:         item.Quality.ID,
					Name:       item.Quality.Name,
					Source:     item.Quality.Source,
					Resolution: item.Quality.Resolution,
					Modifier:   item.Quality.Modifier,
				},
				Allowed: item.Allowed,
				Name:    item.Name,
				ID:      item.ID,
			}
			if len(item.Items) > 0 {
				sonarrItem.Items = make([]structures.SonarrQualityProfileItem, 0, len(item.Items))
				for _, subItem := range item.Items {
					subSonarrItem := structures.SonarrQualityProfileItem{
						Quality: structures.SonarrQuality{

							ID:         subItem.Quality.ID,
							Name:       subItem.Quality.Name,
							Source:     subItem.Quality.Source,
							Resolution: subItem.Quality.Resolution,
							Modifier:   subItem.Quality.Modifier,
						},
						Allowed: subItem.Allowed,
						Name:    subItem.Name,
						ID:      subItem.ID,
					}
					sonarrItem.Items = append(sonarrItem.Items, subSonarrItem)
				}
			}
			sonarrProfile.Items = append(sonarrProfile.Items, sonarrItem)
		}
		for _, formatItem := range profile.FormatItems {
			sonarrProfile.FormatItems = append(sonarrProfile.FormatItems, structures.SonarrFormatItem{
				Format: formatItem.Format,
				Name:   formatItem.Name,
				Score:  formatItem.Score,
			})
		}
		result = append(result, sonarrProfile)
	}

	return ctx.JSON(result)
}
