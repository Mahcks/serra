package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/internal/integrations/emby"
	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/pkg/structures"
)

type LibrarySyncIncrementalJob struct {
	*BaseJob
	embyService emby.Service
}

func NewLibrarySyncIncremental(gctx global.Context, config JobConfig) (*LibrarySyncIncrementalJob, error) {
	// Initialize Emby service
	embyService := emby.New(gctx)

	base := NewBaseJob(gctx, structures.JobLibrarySyncIncremental, config)
	job := &LibrarySyncIncrementalJob{
		BaseJob:     base,
		embyService: embyService,
	}

	return job, nil
}

// Trigger executes the incremental library sync task
func (j *LibrarySyncIncrementalJob) Trigger(ctx context.Context) error {
	return j.Execute(ctx)
}

// Start begins the incremental library sync loop
func (j *LibrarySyncIncrementalJob) Start(ctx context.Context) error {
	slog.Info("Starting incremental library sync", "interval", j.Config().Interval)
	return j.BaseJob.Start(ctx)
}

func (j *LibrarySyncIncrementalJob) Execute(ctx context.Context) error {
	slog.Info("Starting incremental library sync job")

	// Get recently added items from the last 20 minutes (with buffer)
	// This ensures we don't miss any items between runs
	maxAge := time.Now().Add(-20 * time.Minute).Format(time.RFC3339)
	
	// Get recently added items from Emby/Jellyfin
	libraryItems, err := j.embyService.GetRecentlyAddedItems(maxAge)
	if err != nil {
		slog.Error("Failed to fetch recently added items from Emby", "error", err)
		return fmt.Errorf("failed to fetch recently added items: %w", err)
	}

	slog.Info("Fetched recently added items from Emby", "count", len(libraryItems), "since", maxAge)

	// Process recently added items
	insertedCount := 0
	updatedCount := 0
	skippedCount := 0
	
	for _, item := range libraryItems {
		if item.TmdbID == "" {
			skippedCount++
			continue
		}

		// Check if item already exists
		existingItem, err := j.Context().Crate().Sqlite.Query().GetLibraryItemByTMDBID(ctx, sql.NullString{
			String: item.TmdbID,
			Valid:  true,
		})
		
		if err != nil && err != sql.ErrNoRows {
			slog.Error("Failed to check existing library item", 
				"item_id", item.ID,
				"tmdb_id", item.TmdbID,
				"error", err)
			continue
		}

		if err == sql.ErrNoRows {
			// Item doesn't exist, insert it
			err := j.insertLibraryItem(ctx, item)
			if err != nil {
				slog.Error("Failed to insert library item", 
					"item_id", item.ID,
					"name", item.Name,
					"tmdb_id", item.TmdbID,
					"error", err)
				continue
			}
			insertedCount++
		} else {
			// Item exists, update if needed
			if j.shouldUpdateItem(existingItem, item) {
				err := j.updateLibraryItem(ctx, item)
				if err != nil {
					slog.Error("Failed to update library item", 
						"item_id", item.ID,
						"name", item.Name,
						"tmdb_id", item.TmdbID,
						"error", err)
					continue
				}
				updatedCount++
			}
		}
	}

	slog.Info("Incremental library sync completed",
		"total_items", len(libraryItems),
		"inserted", insertedCount,
		"updated", updatedCount,
		"skipped", skippedCount)

	return nil
}

func (j *LibrarySyncIncrementalJob) shouldUpdateItem(existing repository.LibraryItem, new structures.EmbyMediaItem) bool {
	// Compare relevant fields to determine if update is needed
	// For now, just check if the name changed (could be extended)
	return existing.Name != new.Name
}

func (j *LibrarySyncIncrementalJob) insertLibraryItem(ctx context.Context, item structures.EmbyMediaItem) error {
	// Serialize JSON fields
	genresJSON, _ := json.Marshal(item.Genres)
	studiosJSON, _ := json.Marshal(item.Studios)
	peopleJSON, _ := json.Marshal(item.People)
	subtitleTracksJSON, _ := json.Marshal(item.SubtitleTracks)
	audioTracksJSON, _ := json.Marshal(item.AudioTracks)
	backdropImageTagsJSON, _ := json.Marshal(item.BackdropImageTags)
	providerIdsJSON, _ := json.Marshal(item.ProviderIds)
	externalUrlsJSON, _ := json.Marshal(item.ExternalUrls)
	tagsJSON, _ := json.Marshal(item.Tags)
	userDataJSON, _ := json.Marshal(item.UserData)

	// Use direct SQL insert with all fields
	query := `
		INSERT INTO library_items (
			id, name, original_title, type, parent_id, series_id, season_number, episode_number, year, premiere_date, end_date,
			community_rating, critic_rating, official_rating, overview, tagline, genres, studios, people,
			tmdb_id, imdb_id, tvdb_id, musicbrainz_id, path, container, size_bytes, bitrate, width, height,
			aspect_ratio, video_codec, audio_codec, subtitle_tracks, audio_tracks, runtime_ticks, runtime_minutes,
			is_folder, is_resumable, play_count, date_created, date_modified, last_played_date, user_data,
			chapter_images_extracted, primary_image_tag, backdrop_image_tags, logo_image_tag, art_image_tag,
			thumb_image_tag, is_hd, is_4k, is_3d, locked, provider_ids, external_urls, tags, sort_name,
			forced_sort_name, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := j.Context().Crate().Sqlite.DB().ExecContext(ctx, query,
		item.ID, item.Name, nullStringFromString(item.OriginalTitle), item.Type,
		nullStringFromString(item.ParentID), nullStringFromString(item.SeriesID),
		nullInt64FromInt(item.SeasonNumber), nullInt64FromInt(item.EpisodeNumber),
		nullInt64FromInt(item.Year), nullStringFromString(item.PremiereDate), nullStringFromString(item.EndDate),
		nullFloat64FromFloat64(item.CommunityRating), nullFloat64FromFloat64(item.CriticRating),
		nullStringFromString(item.OfficialRating), nullStringFromString(item.Overview), nullStringFromString(item.Tagline),
		nullStringFromBytes(genresJSON), nullStringFromBytes(studiosJSON), nullStringFromBytes(peopleJSON),
		nullStringFromString(item.TmdbID), nullStringFromString(item.ImdbID), nullStringFromString(item.TvdbID),
		nullStringFromString(item.MusicBrainzID), nullStringFromString(item.Path), nullStringFromString(item.Container),
		nullInt64FromInt64(item.SizeBytes), nullInt64FromInt(item.Bitrate), nullInt64FromInt(item.Width), nullInt64FromInt(item.Height),
		nullStringFromString(item.AspectRatio), nullStringFromString(item.VideoCodec), nullStringFromString(item.AudioCodec),
		nullStringFromBytes(subtitleTracksJSON), nullStringFromBytes(audioTracksJSON),
		nullInt64FromInt64(item.RuntimeTicks), nullInt64FromInt(item.RuntimeMinutes),
		item.IsFolder, item.IsResumable, item.PlayCount,
		nullStringFromString(item.DateCreated), nullStringFromString(item.DateModified), nullStringFromString(item.LastPlayedDate),
		nullStringFromBytes(userDataJSON), item.ChapterImagesExtracted,
		nullStringFromString(item.PrimaryImageTag), nullStringFromBytes(backdropImageTagsJSON),
		nullStringFromString(item.LogoImageTag), nullStringFromString(item.ArtImageTag), nullStringFromString(item.ThumbImageTag),
		item.IsHD, item.Is4K, item.Is3D, item.Locked,
		nullStringFromBytes(providerIdsJSON), nullStringFromBytes(externalUrlsJSON), nullStringFromBytes(tagsJSON),
		nullStringFromString(item.SortName), nullStringFromString(item.ForcedSortName), time.Now(), time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert library item: %w", err)
	}

	return nil
}

func (j *LibrarySyncIncrementalJob) updateLibraryItem(ctx context.Context, item structures.EmbyMediaItem) error {
	// For incremental updates, we can use a simple delete and re-insert
	// A more sophisticated approach would be to create an UpdateLibraryItem query
	
	// Delete existing item
	_, err := j.Context().Crate().Sqlite.DB().ExecContext(ctx, 
		"DELETE FROM library_items WHERE tmdb_id = ?", item.TmdbID)
	if err != nil {
		return fmt.Errorf("failed to delete existing library item: %w", err)
	}

	// Insert updated item
	return j.insertLibraryItem(ctx, item)
}

