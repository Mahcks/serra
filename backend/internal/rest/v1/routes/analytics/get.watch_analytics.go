package analytics

import (
	"log/slog"
	"strconv"

	"github.com/mahcks/serra/internal/integrations/jellystat"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

// GetWatchAnalytics returns comprehensive watch analytics from Jellystat
func (rg *RouteGroup) GetWatchAnalytics(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || !user.IsAdmin {
		return apiErrors.ErrInsufficientPermissions()
	}

	// Parse query parameters
	limitParam := ctx.Query("limit", "10")
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit < 1 {
		limit = 10
	}

	// Check if Jellystat is enabled from database settings
	jellystatEnabledStr, err := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingJellystatEnabled.String())
	jellystatEnabled := err == nil && jellystatEnabledStr == "true"

	// Initialize Jellystat service
	jellystatSvc := jellystat.New(rg.gctx)

	// Initialize empty arrays
	libraries := []structures.JellystatLibrary{}
	userActivity := []structures.JellystatUserActivity{}
	popularContent := []structures.JellystatPopularContent{}
	mostViewedContent := []structures.JellystatPopularContent{}
	activeUsers := []structures.JellystatActiveUser{}
	playbackMethods := []structures.JellystatPlaybackMethod{}
	recentlyWatched := []structures.JellystatRecentlyWatched{}

	// Only fetch data if Jellystat is enabled
	if jellystatEnabled {
		// Get library overview
		if libData, err := jellystatSvc.GetLibraryOverview(); err != nil {
			slog.Warn("Failed to get library overview from Jellystat", "error", err)
		} else {
			libraries = libData
		}

		// Get user activity
		if userData, err := jellystatSvc.GetUserActivity(); err != nil {
			slog.Warn("Failed to get user activity from Jellystat", "error", err)
		} else {
			userActivity = userData
		}

		// Get popular content
		if popData, err := jellystatSvc.GetPopularContent(limit); err != nil {
			slog.Warn("Failed to get popular content from Jellystat", "error", err)
		} else {
			popularContent = popData
		}

		// Get most viewed content
		if viewedData, err := jellystatSvc.GetMostViewedContent(limit); err != nil {
			slog.Warn("Failed to get most viewed content from Jellystat", "error", err)
		} else {
			mostViewedContent = viewedData
		}

		// Get most active users (30 days)
		if activeData, err := jellystatSvc.GetMostActiveUsers(30); err != nil {
			slog.Warn("Failed to get active users from Jellystat", "error", err)
		} else {
			activeUsers = activeData
		}

		// Get playback method stats (30 days)
		if playbackData, err := jellystatSvc.GetPlaybackMethodStats(30); err != nil {
			slog.Warn("Failed to get playback method stats from Jellystat", "error", err)
		} else {
			playbackMethods = playbackData
		}

		// Get recently watched
		if recentData, err := jellystatSvc.GetRecentlyWatched(limit); err != nil {
			slog.Warn("Failed to get recently watched from Jellystat", "error", err)
		} else {
			recentlyWatched = recentData
		}
	}

	return ctx.JSON(map[string]interface{}{
		"libraries":           libraries,
		"user_activity":       userActivity,
		"popular_content":     popularContent,
		"most_viewed_content": mostViewedContent,
		"active_users":        activeUsers,
		"playback_methods":    playbackMethods,
		"recently_watched":    recentlyWatched,
		"jellystat_enabled":   jellystatEnabled,
	})
}