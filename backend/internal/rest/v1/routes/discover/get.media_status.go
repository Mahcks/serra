package discover

import (
	"strconv"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

// GetMediaStatus checks if a specific media item is in library or requested by the user
func (rg *RouteGroup) GetMediaStatus(ctx *respond.Ctx) error {
	tmdbIDStr := ctx.Params("tmdb_id")
	mediaType := ctx.Query("media_type")

	// Validate required parameters
	if tmdbIDStr == "" {
		return apiErrors.ErrBadRequest().SetDetail("tmdb_id is required")
	}
	if mediaType == "" {
		return apiErrors.ErrBadRequest().SetDetail("media_type query parameter is required")
	}
	if mediaType != "movie" && mediaType != "tv" {
		return apiErrors.ErrBadRequest().SetDetail("media_type must be 'movie' or 'tv'")
	}

	// Parse TMDB ID
	tmdbID, err := strconv.ParseInt(tmdbIDStr, 10, 64)
	if err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Invalid tmdb_id")
	}

	// Get user from context
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Create a dummy TMDB media item for the enrichment function
	mediaItem := &structures.TMDBMediaItem{
		ID: tmdbID,
	}

	// Use existing enrichment function to get status
	enrichedItem, err := rg.enrichSingleMediaWithStatus(ctx.Context(), mediaItem, user.ID, mediaType)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to check media status: " + err.Error())
	}

	// Return the status information
	return ctx.JSON(map[string]interface{}{
		"tmdb_id":     tmdbID,
		"media_type":  mediaType,
		"in_library":  enrichedItem.InLibrary,
		"requested":   enrichedItem.Requested,
	})
}