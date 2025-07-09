package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/internal/services/auth"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

func (rg *RouteGroup) Logout(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized().SetDetail("No active session found")
	}

	slog.Info("User logging out", "user_id", user.ID, "username", user.Username)

	// Get user from database to determine user type
	dbUser, err := rg.gctx.Crate().Sqlite.Query().GetUserByID(ctx.Context(), user.ID)
	if err != nil {
		slog.Warn("Failed to get user from database during logout", "error", err, "user_id", user.ID)
		// Continue with logout even if we can't get user details
	}

	// Handle logout based on user type
	if dbUser.UserType == "media_server" {
		if err := rg.logoutMediaServerUser(user.ID, user.AccessToken); err != nil {
			slog.Warn("Failed to logout from media server", "error", err, "user_id", user.ID)
			// Continue with local logout even if media server logout fails
		}
	}
	// For local users, no additional logout steps needed

	// Clear the authentication cookie
	ctx.Cookie(rg.gctx.Crate().AuthService.Cookie(auth.CookieAuth, "", time.Second*-1))

	slog.Info("User logged out successfully", "user_id", user.ID, "username", user.Username)
	return ctx.SendStatus(fiber.StatusNoContent)
}

// logoutMediaServerUser handles logout from the media server
func (rg *RouteGroup) logoutMediaServerUser(userID, accessToken string) error {
	// Build logout URL - different endpoints for different media servers
	var logoutURL string
	if rg.Config().MediaServer.Type == structures.ProviderJellyfin {
		logoutURL = fmt.Sprintf("%s/Sessions/Logout", rg.Config().MediaServer.URL)
	} else {
		logoutURL = fmt.Sprintf("%s/Sessions/Logout", rg.Config().MediaServer.URL)
	}

	httpReq, err := http.NewRequest("POST", logoutURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create logout request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Set authorization header for media server
	version := rg.gctx.Bootstrap().Version
	utils.SetMediaServerAuthHeader(httpReq, string(rg.Config().MediaServer.Type), version, accessToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to contact media server for logout: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		slog.Debug("Media server logout returned non-success status", "status", resp.StatusCode)
		// Don't return error for non-success status as session might already be invalid
	}

	return nil
}
