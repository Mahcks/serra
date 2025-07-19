package discover

import (
	"strconv"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

type MediaRatingsResponse struct {
	RottenTomatoes *RottenTomatoesRating `json:"rotten_tomatoes,omitempty"`
	// Future rating services can be added here
	// IMDB           *IMDBRating           `json:"imdb,omitempty"`
	// Metacritic     *MetacriticRating     `json:"metacritic,omitempty"`
}

type RottenTomatoesRating struct {
	Title          string `json:"title"`
	Year           int    `json:"year"`
	Type           string `json:"type"`
	TomatoMeter    int    `json:"tomato_meter"`
	URL            string `json:"url"`
	CriticsRating  string `json:"critics_rating"`
	CriticsScore   int    `json:"critics_score"`
	AudienceRating string `json:"audience_rating"`
	AudienceScore  int    `json:"audience_score"`
}

// GetMediaRatings returns ratings for a media item from various rating services
func (rg *RouteGroup) GetMediaRatings(ctx *respond.Ctx) error {
	tmdbID := ctx.Params("tmdb_id")
	mediaType := ctx.Query("media_type", "movie")
	title := ctx.Query("title")
	yearStr := ctx.Query("year")

	if tmdbID == "" {
		return apiErrors.ErrBadRequest().SetDetail("TMDB ID is required")
	}

	if title == "" {
		return apiErrors.ErrBadRequest().SetDetail("Title is required")
	}

	// Convert year from string to int
	var year int
	if yearStr != "" {
		var err error
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			return apiErrors.ErrBadRequest().SetDetail("Invalid year parameter")
		}
	}

	response := &MediaRatingsResponse{}

	// Get Rotten Tomatoes rating
	if rg.integrations.RottenTomatoes != nil {
		rtRating, err := rg.integrations.RottenTomatoes.GetRottenTomatoesScore(ctx.Context(), title, year, mediaType)
		if err == nil && rtRating != nil {
			response.RottenTomatoes = &RottenTomatoesRating{
				Title:          rtRating.Title,
				Year:           rtRating.Year,
				Type:           rtRating.Type,
				TomatoMeter:    rtRating.TomatoMeter,
				URL:            rtRating.URL,
				CriticsRating:  rtRating.CriticsRating,
				CriticsScore:   rtRating.CriticsScore,
				AudienceRating: rtRating.AudienceRating,
				AudienceScore:  rtRating.AudienceScore,
			}
		}
	}

	// Future rating services can be added here:
	// if rg.integrations.IMDB != nil {
	//     imdbRating, err := rg.integrations.IMDB.GetIMDBRating(ctx.Context(), title, year, mediaType)
	//     if err == nil && imdbRating != nil {
	//         response.IMDB = &IMDBRating{...}
	//     }
	// }

	return ctx.JSON(response)
}