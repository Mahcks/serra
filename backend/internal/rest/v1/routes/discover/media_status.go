package discover

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/pkg/structures"
)

// enrichWithMediaStatus adds request and library status to TMDB media response
func (rg *RouteGroup) enrichWithMediaStatus(ctx context.Context, response *structures.TMDBMediaResponse, userID string, mediaType string) (*structures.TMDBFullMediaResponse, error) {
	if response == nil || len(response.Results) == 0 {
		return &structures.TMDBFullMediaResponse{
			TMDBPageResults: response.TMDBPageResults,
			Results:         []structures.TMDBFullMediaItem{},
		}, nil
	}

	// Extract TMDB IDs for batch processing
	tmdbIDs := make([]int64, 0, len(response.Results))
	tmdbIDStrings := make([]string, 0, len(response.Results))

	for _, item := range response.Results {
		tmdbIDs = append(tmdbIDs, item.ID)
		tmdbIDStrings = append(tmdbIDStrings, strconv.FormatInt(item.ID, 10))
	}

	// Batch check library status for all items - check individually since current implementation doesn't support batch
	libraryMap := make(map[string]bool)
	for _, tmdbIDStr := range tmdbIDStrings {
		inLibrary, err := rg.gctx.Crate().Sqlite.Query().CheckMediaInLibrary(ctx, sql.NullString{
			String: tmdbIDStr,
			Valid:  true,
		})
		if err != nil {
			return nil, err
		}
		libraryMap[tmdbIDStr] = inLibrary
	}

	// For mixed media types (trending), we need to check requests separately for movies and TV
	var requestStatus []repository.CheckMultipleUserRequestsRow
	if mediaType == "mixed" {
		// Separate movies and TV shows
		var movieIDs, tvIDs []int64
		for _, item := range response.Results {
			itemMediaType := item.MediaType
			if itemMediaType == "" {
				// Fallback: determine media type from available fields
				if item.ReleaseDate != "" {
					itemMediaType = "movie"
				} else if item.FirstAirDate != "" {
					itemMediaType = "tv"
				}
			}

			switch itemMediaType {
			case "movie":
				movieIDs = append(movieIDs, item.ID)
			case "tv":
				tvIDs = append(tvIDs, item.ID)
			}
		}

		// Check movie requests individually
		for _, movieID := range movieIDs {
			requested, err := rg.gctx.Crate().Sqlite.Query().CheckUserRequestExists(ctx, repository.CheckUserRequestExistsParams{
				TmdbID: sql.NullInt64{
					Int64: movieID,
					Valid: true,
				},
				MediaType: "movie",
				UserID:    userID,
			})
			if err != nil {
				return nil, err
			}
			if requested {
				requestStatus = append(requestStatus, repository.CheckMultipleUserRequestsRow{
					TmdbID: sql.NullInt64{
						Int64: movieID,
						Valid: true,
					},
					Requested: true,
				})
			}
		}

		// Check TV requests individually
		for _, tvID := range tvIDs {
			requested, err := rg.gctx.Crate().Sqlite.Query().CheckUserRequestExists(ctx, repository.CheckUserRequestExistsParams{
				TmdbID: sql.NullInt64{
					Int64: tvID,
					Valid: true,
				},
				MediaType: "tv",
				UserID:    userID,
			})
			if err != nil {
				return nil, err
			}
			if requested {
				requestStatus = append(requestStatus, repository.CheckMultipleUserRequestsRow{
					TmdbID: sql.NullInt64{
						Int64: tvID,
						Valid: true,
					},
					Requested: true,
				})
			}
		}
	} else {
		// Single media type - check requests individually
		for _, tmdbID := range tmdbIDs {
			requested, err := rg.gctx.Crate().Sqlite.Query().CheckUserRequestExists(ctx, repository.CheckUserRequestExistsParams{
				TmdbID: sql.NullInt64{
					Int64: tmdbID,
					Valid: true,
				},
				MediaType: mediaType,
				UserID:    userID,
			})
			if err != nil {
				return nil, err
			}
			if requested {
				requestStatus = append(requestStatus, repository.CheckMultipleUserRequestsRow{
					TmdbID: sql.NullInt64{
						Int64: tmdbID,
						Valid: true,
					},
					Requested: true,
				})
			}
		}
	}

	// Create request map for quick lookup (libraryMap already created above)

	requestMap := make(map[int64]bool)
	for _, status := range requestStatus {
		if status.TmdbID.Valid {
			requestMap[status.TmdbID.Int64] = status.Requested
		}
	}

	// Build enriched response
	enrichedResults := make([]structures.TMDBFullMediaItem, 0, len(response.Results))
	for _, item := range response.Results {
		tmdbIDStr := strconv.FormatInt(item.ID, 10)

		enrichedItem := structures.TMDBFullMediaItem{
			TMDBMediaItem: item,
			InLibrary:     libraryMap[tmdbIDStr],
			Requested:     requestMap[item.ID],
		}
		enrichedResults = append(enrichedResults, enrichedItem)
	}
	return &structures.TMDBFullMediaResponse{
		TMDBPageResults: response.TMDBPageResults,
		Results:         enrichedResults,
	}, nil
}

// enrichSingleMediaWithStatus adds request and library status to a single TMDB media item
func (rg *RouteGroup) enrichSingleMediaWithStatus(ctx context.Context, item *structures.TMDBMediaItem, userID string, mediaType string) (*structures.TMDBFullMediaItem, error) {
	if item == nil {
		return nil, nil
	}

	tmdbIDStr := strconv.FormatInt(item.ID, 10)

	// Check library status
	inLibrary, err := rg.gctx.Crate().Sqlite.Query().CheckMediaInLibrary(ctx, sql.NullString{
		String: tmdbIDStr,
		Valid:  true,
	})
	if err != nil {
		return nil, err
	}

	// Check request status for user
	requested, err := rg.gctx.Crate().Sqlite.Query().CheckUserRequestExists(ctx, repository.CheckUserRequestExistsParams{
		TmdbID: sql.NullInt64{
			Int64: item.ID,
			Valid: true,
		},
		MediaType: mediaType,
		UserID:    userID,
	})
	if err != nil {
		return nil, err
	}

	return &structures.TMDBFullMediaItem{
		TMDBMediaItem: *item,
		InLibrary:     inLibrary,
		Requested:     requested,
	}, nil
}
