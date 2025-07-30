package users

import (
	"database/sql"
	"log/slog"
	"strings"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

// UpdateUserSettings updates user profile and account settings
func (rg *RouteGroup) UpdateUserSettings(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	var req structures.UpdateUserSettingsRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Invalid request body")
	}

	// Get current user data
	currentUser, err := rg.gctx.Crate().Sqlite.Query().GetUserByID(ctx.Context(), user.ID)
	if err != nil {
		slog.Error("Failed to get current user", "error", err, "user_id", user.ID)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to retrieve user")
	}

	// Update profile if provided
	if req.Profile != nil {
		updateParams := repository.UpdateUserParams{
			ID:        user.ID,
			Username:  currentUser.Username, // Keep current username
			Email:     currentUser.Email,    // Keep current email
			AvatarUrl: currentUser.AvatarUrl, // Keep current avatar
		}

		// Update email if provided and different
		if req.Profile.Email != nil && *req.Profile.Email != "" {
			email := strings.TrimSpace(strings.ToLower(*req.Profile.Email))
			if email != currentUser.Email.String {
				updateParams.Email = sql.NullString{String: email, Valid: true}
			}
		}

		// Update avatar URL if provided
		if req.Profile.AvatarURL != nil {
			avatarURL := strings.TrimSpace(*req.Profile.AvatarURL)
			if avatarURL == "" {
				// Clear avatar URL
				updateParams.AvatarUrl = sql.NullString{Valid: false}
			} else {
				// Set avatar URL
				updateParams.AvatarUrl = sql.NullString{String: avatarURL, Valid: true}
			}
		}

		// Update user profile
		err = rg.gctx.Crate().Sqlite.Query().UpdateUser(ctx.Context(), updateParams)
		if err != nil {
			slog.Error("Failed to update user profile", "error", err, "user_id", user.ID)
			return apiErrors.ErrInternalServerError().SetDetail("Failed to update profile")
		}

		slog.Info("User profile updated", "user_id", user.ID)
	}

	// Update notification preferences if provided
	if req.NotificationPreferences != nil {
		err := rg.gctx.Crate().NotificationService.UpdateUserPreferences(ctx.Context(), user.ID, *req.NotificationPreferences)
		if err != nil {
			slog.Error("Failed to update notification preferences", "error", err, "user_id", user.ID)
			return apiErrors.ErrInternalServerError().SetDetail("Failed to update notification preferences")
		}
		slog.Info("Notification preferences updated", "user_id", user.ID)
	}

	// Update account settings if provided
	if req.AccountSettings != nil {
		settings := req.AccountSettings
		
		if settings.Theme != nil {
			err = rg.gctx.Crate().Sqlite.Query().SetUserSetting(ctx.Context(), repository.SetUserSettingParams{
				UserID: user.ID,
				Key:    "theme",
				Value:  *settings.Theme,
			})
			if err != nil {
				slog.Error("Failed to update theme setting", "error", err, "user_id", user.ID)
				return apiErrors.ErrInternalServerError().SetDetail("Failed to update theme setting")
			}
		}

		if settings.Language != nil {
			err = rg.gctx.Crate().Sqlite.Query().SetUserSetting(ctx.Context(), repository.SetUserSettingParams{
				UserID: user.ID,
				Key:    "language",
				Value:  *settings.Language,
			})
			if err != nil {
				slog.Error("Failed to update language setting", "error", err, "user_id", user.ID)
				return apiErrors.ErrInternalServerError().SetDetail("Failed to update language setting")
			}
		}

		if settings.Timezone != nil {
			err = rg.gctx.Crate().Sqlite.Query().SetUserSetting(ctx.Context(), repository.SetUserSettingParams{
				UserID: user.ID,
				Key:    "timezone",
				Value:  *settings.Timezone,
			})
			if err != nil {
				slog.Error("Failed to update timezone setting", "error", err, "user_id", user.ID)
				return apiErrors.ErrInternalServerError().SetDetail("Failed to update timezone setting")
			}
		}

		if settings.DateFormat != nil {
			err = rg.gctx.Crate().Sqlite.Query().SetUserSetting(ctx.Context(), repository.SetUserSettingParams{
				UserID: user.ID,
				Key:    "date_format",
				Value:  *settings.DateFormat,
			})
			if err != nil {
				slog.Error("Failed to update date format setting", "error", err, "user_id", user.ID)
				return apiErrors.ErrInternalServerError().SetDetail("Failed to update date format setting")
			}
		}

		if settings.TimeFormat != nil {
			err = rg.gctx.Crate().Sqlite.Query().SetUserSetting(ctx.Context(), repository.SetUserSettingParams{
				UserID: user.ID,
				Key:    "time_format",
				Value:  *settings.TimeFormat,
			})
			if err != nil {
				slog.Error("Failed to update time format setting", "error", err, "user_id", user.ID)
				return apiErrors.ErrInternalServerError().SetDetail("Failed to update time format setting")
			}
		}

		slog.Info("Account settings updated", "user_id", user.ID)
	}

	// Update privacy settings if provided
	if req.PrivacySettings != nil {
		privacy := req.PrivacySettings
		
		if privacy.ShowOnlineStatus != nil {
			value := "false"
			if *privacy.ShowOnlineStatus {
				value = "true"
			}
			err = rg.gctx.Crate().Sqlite.Query().SetUserSetting(ctx.Context(), repository.SetUserSettingParams{
				UserID: user.ID,
				Key:    "show_online_status",
				Value:  value,
			})
			if err != nil {
				slog.Error("Failed to update show online status setting", "error", err, "user_id", user.ID)
				return apiErrors.ErrInternalServerError().SetDetail("Failed to update privacy setting")
			}
		}

		if privacy.ShowWatchHistory != nil {
			value := "false"
			if *privacy.ShowWatchHistory {
				value = "true"
			}
			err = rg.gctx.Crate().Sqlite.Query().SetUserSetting(ctx.Context(), repository.SetUserSettingParams{
				UserID: user.ID,
				Key:    "show_watch_history",
				Value:  value,
			})
			if err != nil {
				slog.Error("Failed to update show watch history setting", "error", err, "user_id", user.ID)
				return apiErrors.ErrInternalServerError().SetDetail("Failed to update privacy setting")
			}
		}

		if privacy.ShowRequestHistory != nil {
			value := "false"
			if *privacy.ShowRequestHistory {
				value = "true"
			}
			err = rg.gctx.Crate().Sqlite.Query().SetUserSetting(ctx.Context(), repository.SetUserSettingParams{
				UserID: user.ID,
				Key:    "show_request_history",
				Value:  value,
			})
			if err != nil {
				slog.Error("Failed to update show request history setting", "error", err, "user_id", user.ID)
				return apiErrors.ErrInternalServerError().SetDetail("Failed to update privacy setting")
			}
		}

		slog.Info("Privacy settings updated", "user_id", user.ID)
	}

	// Return success response
	return ctx.JSON(map[string]interface{}{
		"message": "Settings updated successfully",
		"updated": map[string]bool{
			"profile":         req.Profile != nil,
			"notifications":   req.NotificationPreferences != nil,
			"account":         req.AccountSettings != nil,
			"privacy":         req.PrivacySettings != nil,
		},
	})
}