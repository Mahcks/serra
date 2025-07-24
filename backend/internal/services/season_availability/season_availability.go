package season_availability

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/integrations/emby"
	"github.com/mahcks/serra/internal/integrations/tmdb"
	"github.com/mahcks/serra/pkg/structures"
)

type SeasonAvailabilityService struct {
	db         *repository.Queries
	embyClient emby.Service
	tmdbClient tmdb.Service
}

func NewSeasonAvailabilityService(db *repository.Queries, embyClient emby.Service, tmdbClient tmdb.Service) *SeasonAvailabilityService {
	return &SeasonAvailabilityService{
		db:         db,
		embyClient: embyClient,
		tmdbClient: tmdbClient,
	}
}

// SyncShowAvailability syncs availability for a specific show from the media server
func (s *SeasonAvailabilityService) SyncShowAvailability(ctx context.Context, tmdbID int) error {
	if s.embyClient == nil {
		return fmt.Errorf("emby client not configured")
	}

	log.Printf("🔍 Starting sync for TMDB ID %d", tmdbID)

	// Get all episodes for this show from Emby
	episodes, err := s.embyClient.GetEpisodesByTMDB(ctx, tmdbID)
	if err != nil {
		return fmt.Errorf("failed to get episodes from emby: %w", err)
	}

	log.Printf("📺 Found %d episodes for TMDB ID %d", len(episodes), tmdbID)

	// Log first few episodes for debugging
	for i, episode := range episodes {
		if i < 5 { // Log first 5 episodes
			log.Printf("   Episode %d: Season %d, Episode %d, Title: %s", i+1, episode.SeasonNumber, episode.EpisodeNumber, episode.Name)
		}
	}
	if len(episodes) > 5 {
		log.Printf("   ... and %d more episodes", len(episodes)-5)
	}

	// Group episodes by season
	seasonData := make(map[int]seasonInfo)
	for _, episode := range episodes {
		if episode.SeasonNumber == 0 {
			continue // Skip episodes without season number
		}

		season := seasonData[episode.SeasonNumber]
		season.AvailableEpisodes++
		seasonData[episode.SeasonNumber] = season
	}

	log.Printf("📊 Season data for TMDB %d:", tmdbID)
	for seasonNum, info := range seasonData {
		log.Printf("   Season %d: %d episodes", seasonNum, info.AvailableEpisodes)
	}

	// Update database with availability info
	for seasonNumber, info := range seasonData {
		err := s.updateSeasonAvailability(ctx, tmdbID, seasonNumber, info.AvailableEpisodes)
		if err != nil {
			log.Printf("Failed to update season %d availability for TMDB %d: %v", seasonNumber, tmdbID, err)
			continue
		}
	}

	// Update existing requests with new availability
	err = s.updateRequestStatusesFromAvailability(ctx, tmdbID)
	if err != nil {
		log.Printf("Failed to update request statuses for TMDB %d: %v", tmdbID, err)
	}

	return nil
}

// GetSeasonAvailability returns availability info for a specific show
func (s *SeasonAvailabilityService) GetSeasonAvailability(ctx context.Context, tmdbID int) (*structures.ShowAvailability, error) {
	// Get all season availability records for this show
	seasonRecords, err := s.db.GetSeasonAvailabilityByTMDBID(ctx, int64(tmdbID))
	if err != nil {
		return nil, fmt.Errorf("failed to get season availability: %w", err)
	}

	log.Printf("🔍 [GetSeasonAvailability] Found %d season records for TMDB %d", len(seasonRecords), tmdbID)

	// Convert to structures format
	seasons := make([]structures.SeasonAvailability, 0, len(seasonRecords))
	for _, record := range seasonRecords {
		availableEpisodes := int(0)
		if record.AvailableEpisodes.Valid {
			availableEpisodes = int(record.AvailableEpisodes.Int64)
		}

		isComplete := false
		if record.IsComplete.Valid {
			isComplete = record.IsComplete.Bool
		}

		lastUpdated := ""
		if record.LastUpdated.Valid {
			lastUpdated = record.LastUpdated.Time.Format(time.RFC3339)
		}

		episodeCount := int(record.EpisodeCount)
		
		log.Printf("   Season %d: EpisodeCount=%d, AvailableEpisodes=%d, IsComplete=%v (DB values: EpisodeCount=%d, AvailableEpisodes=%v, IsComplete=%v)", 
			int(record.SeasonNumber), episodeCount, availableEpisodes, isComplete,
			record.EpisodeCount, record.AvailableEpisodes, record.IsComplete)

		seasons = append(seasons, structures.SeasonAvailability{
			TmdbID:            int(record.TmdbID),
			SeasonNumber:      int(record.SeasonNumber),
			EpisodeCount:      episodeCount,
			AvailableEpisodes: availableEpisodes,
			IsComplete:        isComplete,
			LastUpdated:       lastUpdated,
		})
	}

	// Calculate overall status
	overallStatus := "not_available"
	if len(seasons) > 0 {
		allComplete := true
		partialAvailable := false
		for _, season := range seasons {
			if season.AvailableEpisodes > 0 {
				partialAvailable = true
			}
			if !season.IsComplete {
				allComplete = false
			}
		}

		if allComplete {
			overallStatus = "complete"
		} else if partialAvailable {
			overallStatus = "partial"
		}
	}

	return &structures.ShowAvailability{
		TmdbID:        tmdbID,
		TotalSeasons:  len(seasons),
		Seasons:       seasons,
		OverallStatus: overallStatus,
	}, nil
}

// UpdateSeasonFromMediaServer updates a single season's availability from media server
func (s *SeasonAvailabilityService) UpdateSeasonFromMediaServer(ctx context.Context, tmdbID int, seasonNumber int, totalEpisodes int) error {
	if s.embyClient == nil {
		return fmt.Errorf("emby client not configured")
	}

	// Get episodes for this specific season
	episodes, err := s.embyClient.GetEpisodesByTMDBAndSeason(ctx, tmdbID, seasonNumber)
	if err != nil {
		return fmt.Errorf("failed to get episodes for season %d: %w", seasonNumber, err)
	}

	availableEpisodes := len(episodes)

	// Update database
	err = s.upsertSeasonAvailability(ctx, tmdbID, seasonNumber, totalEpisodes, availableEpisodes)
	if err != nil {
		return fmt.Errorf("failed to update season availability: %w", err)
	}

	// Update request statuses
	err = s.updateSeasonRequestStatus(ctx, tmdbID, seasonNumber, availableEpisodes, totalEpisodes)
	if err != nil {
		log.Printf("Failed to update request status for season %d: %v", seasonNumber, err)
	}

	return nil
}

// Private helper methods

type seasonInfo struct {
	AvailableEpisodes int
	TotalEpisodes     int
}

func (s *SeasonAvailabilityService) updateSeasonAvailability(ctx context.Context, tmdbID int, seasonNumber int, availableEpisodes int) error {
	log.Printf("🔧 Starting season availability update for TMDB %d, Season %d with %d available episodes", 
		tmdbID, seasonNumber, availableEpisodes)
	
	// First, check if we already have a record for this season to preserve the total episode count
	existingRecord, err := s.db.GetSeasonAvailabilityByTMDBIDAndSeason(ctx, repository.GetSeasonAvailabilityByTMDBIDAndSeasonParams{
		TmdbID:       int64(tmdbID),
		SeasonNumber: int64(seasonNumber),
	})
	
	var totalEpisodes int64 = int64(availableEpisodes) // Default fallback
	
	if err == nil {
		// Use existing total episode count if we have it
		totalEpisodes = existingRecord.EpisodeCount
		log.Printf("📊 Found existing record: TotalEpisodes=%d, will update with AvailableEpisodes=%d", 
			totalEpisodes, availableEpisodes)
	} else if err != sql.ErrNoRows {
		log.Printf("⚠️ Error getting existing record (will continue): %v", err)
	} else {
		// No existing record found, try to get correct episode count from TMDB
		if s.tmdbClient != nil {
			if tmdbEpisodeCount, err := s.getTotalEpisodesFromTMDB(ctx, tmdbID, seasonNumber); err == nil && tmdbEpisodeCount > 0 {
				totalEpisodes = int64(tmdbEpisodeCount)
				log.Printf("📡 Got total episodes from TMDB: Season %d has %d total episodes", seasonNumber, totalEpisodes)
			} else {
				log.Printf("⚠️ Failed to get episode count from TMDB (will use fallback): %v", err)
			}
		}
		log.Printf("🆕 No existing record found, creating new one with TotalEpisodes=%d", totalEpisodes)
	}
	
	// Calculate if season is complete
	isComplete := totalEpisodes > 0 && int64(availableEpisodes) >= totalEpisodes
	
	log.Printf("🔧 Updating season %d for TMDB %d: TotalEpisodes=%d, AvailableEpisodes=%d, IsComplete=%v", 
		seasonNumber, tmdbID, totalEpisodes, availableEpisodes, isComplete)
	
	err = s.db.UpsertSeasonAvailability(ctx, repository.UpsertSeasonAvailabilityParams{
		TmdbID:            int64(tmdbID),
		SeasonNumber:      int64(seasonNumber),
		EpisodeCount:      totalEpisodes, // Preserve existing total or use fallback
		AvailableEpisodes: sql.NullInt64{Int64: int64(availableEpisodes), Valid: true},
		IsComplete:        sql.NullBool{Bool: isComplete, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to upsert season availability: %w", err)
	}
	
	log.Printf("✅ Successfully updated season availability for TMDB %d, Season %d", tmdbID, seasonNumber)
	return nil
}

func (s *SeasonAvailabilityService) upsertSeasonAvailability(ctx context.Context, tmdbID int, seasonNumber int, totalEpisodes int, availableEpisodes int) error {
	isComplete := totalEpisodes > 0 && availableEpisodes >= totalEpisodes

	err := s.db.UpsertSeasonAvailability(ctx, repository.UpsertSeasonAvailabilityParams{
		TmdbID:            int64(tmdbID),
		SeasonNumber:      int64(seasonNumber),
		EpisodeCount:      int64(totalEpisodes),
		AvailableEpisodes: sql.NullInt64{Int64: int64(availableEpisodes), Valid: true},
		IsComplete:        sql.NullBool{Bool: isComplete, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to upsert season availability: %w", err)
	}
	return nil
}

func (s *SeasonAvailabilityService) updateRequestStatusesFromAvailability(ctx context.Context, tmdbID int) error {
	// For now, just log that we would update request statuses
	// This will be implemented after we add the missing database queries
	log.Printf("Would update request statuses for TMDB %d based on availability", tmdbID)
	return nil
}

func (s *SeasonAvailabilityService) getSeasonAvailability(ctx context.Context, tmdbID int, seasonNumber int) (*structures.SeasonAvailability, error) {
	record, err := s.db.GetSeasonAvailabilityByTMDBIDAndSeason(ctx, repository.GetSeasonAvailabilityByTMDBIDAndSeasonParams{
		TmdbID:       int64(tmdbID),
		SeasonNumber: int64(seasonNumber),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			// Return default if not found
			return &structures.SeasonAvailability{
				TmdbID:            tmdbID,
				SeasonNumber:      seasonNumber,
				EpisodeCount:      0,
				AvailableEpisodes: 0,
				IsComplete:        false,
				LastUpdated:       time.Now().Format(time.RFC3339),
			}, nil
		}
		return nil, fmt.Errorf("failed to get season availability: %w", err)
	}

	availableEpisodes := int(0)
	if record.AvailableEpisodes.Valid {
		availableEpisodes = int(record.AvailableEpisodes.Int64)
	}

	isComplete := false
	if record.IsComplete.Valid {
		isComplete = record.IsComplete.Bool
	}

	lastUpdated := ""
	if record.LastUpdated.Valid {
		lastUpdated = record.LastUpdated.Time.Format(time.RFC3339)
	}

	return &structures.SeasonAvailability{
		TmdbID:            int(record.TmdbID),
		SeasonNumber:      int(record.SeasonNumber),
		EpisodeCount:      int(record.EpisodeCount),
		AvailableEpisodes: availableEpisodes,
		IsComplete:        isComplete,
		LastUpdated:       lastUpdated,
	}, nil
}

func (s *SeasonAvailabilityService) updateSeasonRequestStatus(ctx context.Context, tmdbID int, seasonNumber int, availableEpisodes int, totalEpisodes int) error {
	// This method updates the status of specific season requests
	// For now, we'll delegate to the more comprehensive updateRequestStatusesFromAvailability
	return s.updateRequestStatusesFromAvailability(ctx, tmdbID)
}

// getTotalEpisodesFromTMDB gets the correct total episode count for a season from TMDB
func (s *SeasonAvailabilityService) getTotalEpisodesFromTMDB(ctx context.Context, tmdbID int, seasonNumber int) (int, error) {
	if s.tmdbClient == nil {
		return 0, fmt.Errorf("TMDB client not available")
	}
	
	seasonDetails, err := s.tmdbClient.GetSeasonDetails(strconv.Itoa(tmdbID), strconv.Itoa(seasonNumber))
	if err != nil {
		return 0, fmt.Errorf("failed to get season details from TMDB: %w", err)
	}
	
	return len(seasonDetails.Episodes), nil
}
