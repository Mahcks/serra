package users

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

// GetUserAvatar proxies avatar images from the media server
func (rg *RouteGroup) GetUserAvatar(ctx *respond.Ctx) error {
	userID := ctx.Params("id")
	if userID == "" {
		return apiErrors.ErrBadRequest().SetDetail("user ID is required")
	}

	// Get user to verify they exist and get their access token
	user, err := rg.gctx.Crate().Sqlite.Query().GetUserByID(ctx.Context(), userID)
	if err != nil {
		return apiErrors.ErrNotFound().SetDetail("user not found")
	}

	// Check user type - local users don't have media server avatars
	if user.UserType == "local" {
		return apiErrors.ErrNotFound().SetDetail("local users don't have avatars yet")
	}

	// Check if user has an access token (needed for media server auth)
	if !user.AccessToken.Valid || user.AccessToken.String == "" {
		return apiErrors.ErrNotFound().SetDetail("user avatar not available")
	}

	// Build media server avatar URL
	mediaServerURL := fmt.Sprintf(
		"%s/Users/%s/Images/Primary",
		rg.gctx.Crate().Config.Get().MediaServer.URL.String(),
		userID,
	)

	// Create request to media server
	req, err := http.NewRequest("GET", mediaServerURL, nil)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to create media server request")
	}

	// Add authentication headers
	authKeyHeader := utils.Ternary(
		rg.gctx.Crate().Config.Get().MediaServer.Type == structures.ProviderEmby,
		"X-Emby-Token",
		"X-Jellyfin-Token",
	)
	req.Header.Set(authKeyHeader, user.AccessToken.String)

	// Make request to media server
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to fetch avatar from media server")
	}
	defer resp.Body.Close()

	// Check if avatar exists
	if resp.StatusCode == http.StatusNotFound {
		return apiErrors.ErrNotFound().SetDetail("user avatar not found")
	}
	if resp.StatusCode != http.StatusOK {
		return apiErrors.ErrInternalServerError().SetDetail("media server error")
	}

	// Set appropriate headers
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg" // Default fallback
	}

	ctx.Set("Content-Type", contentType)
	ctx.Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	ctx.Set("Access-Control-Allow-Origin", "*")
	ctx.Set("Access-Control-Allow-Methods", "GET")
	ctx.Set("Access-Control-Allow-Headers", "Content-Type")

	// Stream the image data
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to read avatar data")
	}

	return ctx.Send(imageData)
}
