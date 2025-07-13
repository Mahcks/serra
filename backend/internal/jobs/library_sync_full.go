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
	"github.com/mahcks/serra/pkg/structures"
)

type LibrarySyncFullJob struct {
	*BaseJob
	embyService emby.Service
}

func NewLibrarySyncFull(gctx global.Context, config JobConfig) (*LibrarySyncFullJob, error) {
	// Initialize Emby service
	embyService := emby.New(gctx)

	base := NewBaseJob(gctx, structures.JobLibrarySyncFull, config)
	job := &LibrarySyncFullJob{
		BaseJob:     base,
		embyService: embyService,
	}

	return job, nil
}

// Trigger executes the full library sync task
func (j *LibrarySyncFullJob) Trigger(ctx context.Context) error {
	return j.Execute(ctx)
}

// Start begins the full library sync loop
func (j *LibrarySyncFullJob) Start(ctx context.Context) error {
	slog.Info("Starting full library sync", "interval", j.Config().Interval)
	return j.BaseJob.Start(ctx)
}

func (j *LibrarySyncFullJob) Execute(ctx context.Context) error {
	slog.Info("Starting full library sync job")

	// Get all library items from Emby/Jellyfin
	libraryItems, err := j.embyService.GetAllLibraryItems()
	if err != nil {
		slog.Error("Failed to fetch library items from Emby", "error", err)
		return fmt.Errorf("failed to fetch library items: %w", err)
	}

	slog.Info("Fetched library items from Emby", "count", len(libraryItems))

	// Clear existing library data for full sync
	if err := j.clearLibraryData(ctx); err != nil {
		slog.Error("Failed to clear existing library data", "error", err)
		return fmt.Errorf("failed to clear library data: %w", err)
	}

	// Insert new library data
	insertedCount := 0
	skippedCount := 0
	
	for _, item := range libraryItems {
		if item.TmdbID == "" {
			skippedCount++
			continue
		}

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
	}

	slog.Info("Full library sync completed",
		"total_items", len(libraryItems),
		"inserted", insertedCount,
		"skipped", skippedCount)

	return nil
}

func (j *LibrarySyncFullJob) clearLibraryData(ctx context.Context) error {
	// Delete all existing library items
	_, err := j.Context().Crate().Sqlite.DB().ExecContext(ctx, "DELETE FROM library_items")
	if err != nil {
		return fmt.Errorf("failed to clear library data: %w", err)
	}
	
	slog.Debug("Cleared existing library data")
	return nil
}

func (j *LibrarySyncFullJob) insertLibraryItem(ctx context.Context, item structures.EmbyMediaItem) error {
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

	// Use direct SQL insert with all fields since CreateLibraryItemFull has issues
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

// Helper functions for null handling
func nullStringFromString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullStringFromBytes(b []byte) sql.NullString {
	if len(b) == 0 || string(b) == "null" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: string(b), Valid: true}
}

func nullInt64FromInt(i int) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(i), Valid: i > 0}
}

func nullInt64FromInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: i > 0}
}

func nullFloat64FromFloat64(f float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: f, Valid: f > 0}
}