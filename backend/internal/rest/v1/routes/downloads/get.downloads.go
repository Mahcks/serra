package downloads

import (
	"log/slog"
	"time"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

func (rg *RouteGroup) GetDownloads(ctx *respond.Ctx) error {
	// Get the downloads from the database
	downloads, err := rg.gctx.Crate().Sqlite.Query().ListDownloads(ctx.Context())
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail(err.Error())
	}

	slog.Debug("GetDownloads: Retrieved from database", "count", len(downloads))

	var result []structures.Download

	for _, d := range downloads {
		slog.Debug("GetDownloads: Processing download",
			"id", d.ID,
			"title", d.Title,
			"progress", d.Progress.Float64,
			"status_valid", d.Status.Valid,
			"status", d.Status.String,
			"timeLeft_valid", d.TimeLeft.Valid,
			"timeLeft", d.TimeLeft.String)

		var tmdbID *int64
		if d.TmdbID.Valid {
			tmdbID = &d.TmdbID.Int64
		}

		var tvdbID *int64
		if d.TvdbID.Valid {
			tvdbID = &d.TvdbID.Int64
		}

		var hash *string
		if d.Hash.Valid {
			hash = &d.Hash.String
		}

		var timeLeft *string
		if d.TimeLeft.Valid {
			timeLeft = &d.TimeLeft.String
		}

		var status *string
		if d.Status.Valid {
			status = &d.Status.String
		}

		var updatedAt *string
		if d.LastUpdated.Valid {
			formatted := d.LastUpdated.Time.Format(time.RFC3339)
			updatedAt = &formatted
		}

		download := structures.Download{
			ID:           d.ID,
			Title:        d.Title,
			TorrentTitle: d.TorrentTitle,
			Source:       d.Source,
			TmdbID:       tmdbID,
			TvdbID:       tvdbID,
			Hash:         hash,
			Progress:     d.Progress.Float64,
			TimeLeft:     timeLeft,
			Status:       status,
			UpdatedAt:    updatedAt,
		}

		slog.Debug("GetDownloads: Created download struct",
			"id", download.ID,
			"title", download.Title,
			"progress", download.Progress,
			"status", download.Status,
			"timeLeft", download.TimeLeft)

		result = append(result, download)
	}

	slog.Debug("GetDownloads: Returning result", "count", len(result))

	// Return the downloads as a JSON response
	return ctx.JSON(result)
}
