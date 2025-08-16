package discover

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

func (rg *RouteGroup) GetMediaDetails(ctx *respond.Ctx) error {
	id := ctx.Params("id")
	mediaType := ctx.Query("type") // "movie" or "tv"

	if id == "" || (mediaType != "movie" && mediaType != "tv") {
		return apiErrors.ErrBadRequest().SetDetail("Missing or invalid id/type")
	}

	// Embedded TMDB API key
	tmdbAPIKey := "eyJhbGciOiJIUzI1NiJ9.eyJhdWQiOiJlNzc2YzY0MWFjZTEyM2UwYTJmNzhlMGM3MjU5N2M1ZSIsIm5iZiI6MTcyNjMyOTg2Ni44ODksInN1YiI6IjY2ZTQxNDBhMDAwMDAwMDAwMGJkMDc3NyIsInNjb3BlcyI6WyJhcGlfcmVhZCJdLCJ2ZXJzaW9uIjoxfQ.3HV9FmPHUk5lX2xWuKU_SZSklsKI1QTHD5d76eKz5pI"
	
	url := fmt.Sprintf(
		"https://api.themoviedb.org/3/%s/%s?language=en-US&append_to_response=videos,credits&api_key=%s",
		mediaType, id, tmdbAPIKey,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail(fmt.Sprintf("failed to build request: %s", err))
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail(fmt.Sprintf("failed to fetch details: %s", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return apiErrors.ErrInternalServerError().SetDetail(fmt.Sprintf("TMDB API returned %d", resp.StatusCode))
	}

	switch mediaType {
	case "tv":
		var result structures.TVDetails
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return apiErrors.ErrInternalServerError().SetDetail(fmt.Sprintf("failed to decode TMDB TV response: %s", err))
		}
		return ctx.JSON(result)
	case "movie":
		var result structures.MovieDetails
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return apiErrors.ErrInternalServerError().SetDetail(fmt.Sprintf("failed to decode TMDB Movie response: %s", err))
		}
		return ctx.JSON(result)
	}

	return apiErrors.ErrBadRequest().SetDetail("Unsupported media type")
}
