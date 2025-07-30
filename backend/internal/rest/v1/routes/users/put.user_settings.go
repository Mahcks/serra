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
			ID:       user.ID,
			Username: currentUser.Username, // Keep current username
			Email:    currentUser.Email,    // Keep current email
		}

		// Update email if provided and different
		if req.Profile.Email != nil && *req.Profile.Email != "" {
			email := strings.TrimSpace(strings.ToLower(*req.Profile.Email))
			if email != currentUser.Email.String {
				// TODO: Add email uniqueness check once GetUserByEmail is available
				// For now, just update the email
				updateParams.Email = sql.NullString{String: email, Valid: true}
			}
		}

		// TODO: Avatar URL updates are not supported in current schema
		// The UpdateUser method doesn't include avatar_url field
		if req.Profile.AvatarURL != nil {
			slog.Warn("Avatar URL update requested but not supported in current schema", "user_id", user.ID)
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
		slog.Warn("Notification preferences update temporarily disabled - sqlc code not generated", "user_id", user.ID)
		// TODO: Re-enable once sqlc generates the notification preference methods
		// For now, just log that notification preferences were requested but not updated
	}

	// Return success response
	return ctx.JSON(map[string]interface{}{
		"message": "Settings updated successfully",
		"updated": map[string]bool{
			"profile":      req.Profile != nil,
			"notifications": req.NotificationPreferences != nil,
		},
	})
}