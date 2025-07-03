package downloads

import (
	"log/slog"
	"time"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

func (rg *RouteGroup) GetDownloads(ctx *respond.Ctx) error {
	// Get the downloads from the database
	downloads, err := rg.gctx.Crate().Sqlite.Query().ListDownloads(ctx.Context())
	if err != nil {
		slog.Error("GetDownloads: Failed to retrieve downloads from database", "error", err)
		return apiErrors.ErrInternalServerError().SetDetail(err.Error())
	}

	// Always initialize result as empty array, even if no downloads
	var result []structures.Download

	// If no downloads found, return empty array
	if len(downloads) == 0 {
		return ctx.JSON(result)
	}

	for _, d := range downloads {
		download := structures.Download{
			ID:           d.ID,
			Title:        d.Title,
			TorrentTitle: d.TorrentTitle,
			Source:       d.Source,
			TmdbID:       utils.NullableInt64{NullInt64: d.TmdbID}.ToPointer(),
			TvdbID:       utils.NullableInt64{NullInt64: d.TvdbID}.ToPointer(),
			Hash:         utils.NullableString{NullString: d.Hash}.ToPointer(),
			Progress:     utils.NullableFloat64{NullFloat64: d.Progress}.Or(0.0),
			TimeLeft:     utils.NullableString{NullString: d.TimeLeft}.ToPointer(),
			Status:       utils.NullableString{NullString: d.Status}.ToPointer(),
			UpdatedAt:    nil,
		}

		// Handle LastUpdated using wrapper
		lastUpdated := utils.NullableTime{NullTime: d.LastUpdated}
		if lastUpdated.IsValid() {
			formatted := lastUpdated.Or(time.Time{}).Format(time.RFC3339)
			download.UpdatedAt = &formatted
		}

		result = append(result, download)
	}

	slog.Debug("GetDownloads: Returning result", "count", len(result))

	// Return the downloads as a JSON response
	return ctx.JSON(result)
}
