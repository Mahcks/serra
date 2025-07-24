package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/internal/services/auth"
	"github.com/mahcks/serra/internal/services"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

// checkAuthMethodEnabled checks if a specific authentication method is enabled in settings
func (rg *RouteGroup) checkAuthMethodEnabled(ctx *respond.Ctx, settingKey string) (bool, error) {
	setting, err := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), settingKey)
	if err != nil {
		// If setting doesn't exist, use defaults
		if settingKey == structures.SettingEnableMediaServerAuth.String() {
			return true, nil // Default: enabled
		}
		return false, nil // Default: disabled for local auth
	}
	
	// Handle empty settings with defaults
	if setting == "" {
		if settingKey == structures.SettingEnableMediaServerAuth.String() {
			return true, nil // Default: enabled
		}
		return false, nil // Default: disabled for local auth
	}
	
	return setting == "true", nil
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	User struct {
		ID              string `json:"Id"`
		Username        string `json:"Name"`
		PrimaryImageTag string `json:"PrimaryImageTag,omitempty"`
	} `json:"User"`
	Accesstoken string `json:"AccessToken"`
}

func (rg *RouteGroup) Authenticate(ctx *respond.Ctx) error {
	var req AuthRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Failed to parse request body")
	}

	// Check if media server authentication is enabled
	mediaServerAuthEnabled, err := rg.checkAuthMethodEnabled(ctx, structures.SettingEnableMediaServerAuth.String())
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to check authentication settings")
	}
	if !mediaServerAuthEnabled {
		return apiErrors.ErrForbidden().SetDetail("Media server authentication is disabled")
	}

	// Authenticate with media server
	mediaServerResponse, err := rg.authenticateWithMediaServer(req.Username, req.Password)
	if err != nil {
		return err
	}

	// Check if this is the first user (before creating the user)
	allUsers, err := rg.gctx.Crate().Sqlite.Query().GetAllUsers(ctx.Context())
	if err != nil {
		slog.Error("Failed to check existing users", "error", err)
		return apiErrors.ErrInternalServerError().SetDetail("failed to check existing users")
	}
	isFirstUser := len(allUsers) == 0

	// Store user in database
	user, err := rg.storeMediaServerUser(ctx, mediaServerResponse)
	if err != nil {
		return err
	}

	// Assign owner permission to first user
	if isFirstUser {
		if err := rg.assignOwnerPermission(ctx, user); err != nil {
			slog.Error("Failed to assign owner permission to first user", "error", err, "user_id", user.ID, "username", user.Username)
			// Don't fail the login, but log the error
		} else {
			slog.Info("Assigned owner permission to first user", "user_id", user.ID, "username", user.Username)
		}
	}

	// Create JWT token and set cookie
	token, _, err := rg.gctx.Crate().AuthService.CreateAccessToken(user.ID, user.Username, mediaServerResponse.Accesstoken, true)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to create JWT token")
	}

	ctx.Cookie(rg.gctx.Crate().AuthService.Cookie(auth.CookieAuth, token, time.Hour*24*14))
	return ctx.SendStatus(fiber.StatusNoContent)
}

// authenticateWithMediaServer handles the media server authentication request
func (rg *RouteGroup) authenticateWithMediaServer(username, password string) (*authResponse, error) {
	authPayload := map[string]interface{}{
		"Username": username,
		"Pw":       password,
	}

	authPayloadBytes, err := json.Marshal(authPayload)
	if err != nil {
		return nil, apiErrors.ErrInternalServerError().SetDetail("Failed to create authentication request")
	}

	mediaServerURL := fmt.Sprintf("%s/Users/AuthenticateByName", rg.Config().MediaServer.URL)
	httpReq, err := http.NewRequest("POST", mediaServerURL, bytes.NewBuffer(authPayloadBytes))
	if err != nil {
		return nil, apiErrors.ErrInternalServerError().SetDetail("Failed to create media server request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	utils.SetMediaServerAuthHeaderForAuth(httpReq, string(rg.Config().MediaServer.Type), rg.gctx.Bootstrap().Version)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, apiErrors.ErrInternalServerError().SetDetail("failed to contact media server")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		slog.Debug("Media server authentication failed", "status", resp.StatusCode, "response", string(respBody))
		return nil, apiErrors.ErrUnauthorized().SetDetail("Media server rejected credentials: " + string(respBody))
	}

	var res authResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, apiErrors.ErrInternalServerError().SetDetail("Failed to decode media server response")
	}

	return &res, nil
}

// storeMediaServerUser creates or updates a media server user in the database
func (rg *RouteGroup) storeMediaServerUser(ctx *respond.Ctx, mediaServerResponse *authResponse) (*repository.User, error) {
	var avatarURL string
	if mediaServerResponse.User.PrimaryImageTag != "" {
		avatarURL = fmt.Sprintf("/users/%s/avatar", mediaServerResponse.User.ID)
	}

	user, err := rg.gctx.Crate().Sqlite.Query().CreateUser(ctx.Context(), repository.CreateUserParams{
		ID:           mediaServerResponse.User.ID,
		Username:     mediaServerResponse.User.Username,
		AccessToken:  utils.NewNullString(mediaServerResponse.Accesstoken),
		Email:        utils.NewNullString(""),
		AvatarUrl:    utils.NewNullString(avatarURL),
		UserType:     "media_server",
		PasswordHash: utils.NewNullString(""),
	})
	if err != nil {
		slog.Debug("failed to store user", "error", err)
		return nil, apiErrors.ErrInternalServerError().SetDetail("failed to store user in database")
	}

	// Assign default permissions to new user using dynamic service
	defaultPermissionsService := services.NewDynamicDefaultPermissionsService(rg.gctx.Crate().Sqlite.Query())
	if err := defaultPermissionsService.AssignDefaultPermissions(ctx.Context(), user.ID); err != nil {
		slog.Error("Failed to assign default permissions to new user", "error", err, "user_id", user.ID, "username", user.Username)
		// Don't fail the user creation, but log the error
	} else {
		slog.Info("Assigned default permissions to new user", "user_id", user.ID, "username", user.Username)
	}

	return &user, nil
}

// assignOwnerPermission assigns owner permission to a user
func (rg *RouteGroup) assignOwnerPermission(ctx *respond.Ctx, user *repository.User) error {
	return rg.gctx.Crate().Sqlite.Query().AssignUserPermission(ctx.Context(), repository.AssignUserPermissionParams{
		UserID:       user.ID,
		PermissionID: permissions.Owner,
	})
}
