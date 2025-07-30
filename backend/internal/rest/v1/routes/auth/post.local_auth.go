package auth

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

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


// AuthenticateLocal handles authentication for both media server and local users
func (rg *RouteGroup) AuthenticateLocal(ctx *respond.Ctx) error {
	var req AuthRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Failed to parse request body")
	}

	slog.Info("Login attempt", "username", req.Username, "normalized_username", strings.ToLower(req.Username))

	// First, try local user authentication (case-insensitive)
	localUser, err := rg.gctx.Crate().Sqlite.Query().GetUserByUsername(ctx.Context(), strings.ToLower(req.Username))
	if err == nil {
		slog.Info("Local user found", "user_id", localUser.ID, "username", localUser.Username, "has_password", localUser.PasswordHash.Valid)
		
		// Check if local authentication is enabled
		localAuthEnabled, err := rg.checkAuthMethodEnabled(ctx, structures.SettingEnableLocalAuth.String())
		if err != nil {
			slog.Error("Failed to check auth method", "error", err)
			return apiErrors.ErrInternalServerError().SetDetail("Failed to check authentication settings")
		}
		slog.Info("Local auth enabled check", "enabled", localAuthEnabled)
		
		if !localAuthEnabled {
			return apiErrors.ErrForbidden().SetDetail("Local authentication is disabled")
		}

		// Local user found, verify password
		slog.Info("Verifying password", "password_hash_length", len(localUser.PasswordHash.String))
		if err := bcrypt.CompareHashAndPassword([]byte(localUser.PasswordHash.String), []byte(req.Password)); err != nil {
			slog.Error("Password verification failed", "error", err)
			return apiErrors.ErrUnauthorized().SetDetail("Invalid credentials")
		}
		slog.Info("Password verification successful")

		// Create JWT token for local user
		token, _, err := rg.gctx.Crate().AuthService.CreateAccessToken(localUser.ID, localUser.Username, "", false) // Local users aren't admin by default
		if err != nil {
			return apiErrors.ErrInternalServerError().SetDetail("failed to create JWT token")
		}

		// Store the token in a cookie
		ctx.Cookie(rg.gctx.Crate().AuthService.Cookie(auth.CookieAuth, token, time.Hour*24*14))

		return ctx.SendStatus(fiber.StatusNoContent)
	} else {
		slog.Info("Local user not found", "error", err, "username", strings.ToLower(req.Username))
	}

	// If no local user found, check if media server authentication is enabled
	mediaServerAuthEnabled, err := rg.checkAuthMethodEnabled(ctx, structures.SettingEnableMediaServerAuth.String())
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to check authentication settings")
	}
	if !mediaServerAuthEnabled {
		return apiErrors.ErrForbidden().SetDetail("Media server authentication is disabled")
	}

	// Try media server authentication (fallback to original flow)
	return rg.Authenticate(ctx)
}

// RegisterLocalUser creates a new local user account
func (rg *RouteGroup) RegisterLocalUser(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Check if user has owner OR admin.users permission to create users
	hasOwnerPermission, err := rg.gctx.Crate().Sqlite.Query().CheckUserPermission(ctx.Context(), repository.CheckUserPermissionParams{
		UserID:       user.ID,
		PermissionID: "owner",
	})
	if err != nil {
		hasOwnerPermission = false
	}

	hasAdminUsersPermission, err := rg.gctx.Crate().Sqlite.Query().CheckUserPermission(ctx.Context(), repository.CheckUserPermissionParams{
		UserID:       user.ID,
		PermissionID: "admin.users",
	})
	if err != nil {
		hasAdminUsersPermission = false
	}

	if !hasOwnerPermission && !hasAdminUsersPermission {
		return apiErrors.ErrForbidden().SetDetail("Missing required permission: owner or admin.users")
	}

	var req structures.LocalUserRegistrationRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Invalid request body")
	}

	// Check if username already exists (case-insensitive)
	normalizedUsername := strings.ToLower(req.Username)
	_, err = rg.gctx.Crate().Sqlite.Query().GetUserByUsername(ctx.Context(), normalizedUsername)
	if err == nil {
		return apiErrors.ErrBadRequest().SetDetail("Username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to hash password")
	}

	// Generate unique ID for local user
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to generate user ID")
	}
	userID := hex.EncodeToString(randomBytes)

	// Create local user
	newUser, err := rg.gctx.Crate().Sqlite.Query().CreateLocalUser(ctx.Context(), repository.CreateLocalUserParams{
		ID:           userID,
		Username:     normalizedUsername, // Store lowercase username
		Email:        utils.NewNullString(req.Email),
		PasswordHash: utils.NewNullString(string(hashedPassword)),
		AvatarUrl:    utils.NewNullString(""), // Local users start without avatars
	})
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to create user")
	}

	// Assign permissions to new local user
	if len(req.Permissions) > 0 {
		// Use specific permissions if provided
		for _, permission := range req.Permissions {
			// Validate permission
			if !permissions.IsValidPermission(permission) {
				slog.Warn("Invalid permission requested for new user", "permission", permission, "user_id", newUser.ID)
				continue
			}
			
			err := rg.gctx.Crate().Sqlite.Query().AssignUserPermission(ctx.Context(), repository.AssignUserPermissionParams{
				UserID:       newUser.ID,
				PermissionID: permission,
			})
			if err != nil {
				slog.Error("Failed to assign permission to new local user", "error", err, "permission", permission, "user_id", newUser.ID)
			}
		}
		slog.Info("Assigned custom permissions to new local user", "user_id", newUser.ID, "username", newUser.Username, "permissions", req.Permissions)
	} else {
		// Use default permissions if no specific permissions provided
		defaultPermissionsService := services.NewDynamicDefaultPermissionsService(rg.gctx.Crate().Sqlite.Query())
		if err := defaultPermissionsService.AssignDefaultPermissions(ctx.Context(), newUser.ID); err != nil {
			slog.Error("Failed to assign default permissions to new local user", "error", err, "user_id", newUser.ID, "username", newUser.Username)
			// Don't fail the user creation, but log the error
		} else {
			slog.Info("Assigned default permissions to new local user", "user_id", newUser.ID, "username", newUser.Username)
		}
	}

	// Create default notification preferences for new user
	if err := rg.gctx.Crate().NotificationService.CreateDefaultPreferencesForUser(ctx.Context(), newUser.ID); err != nil {
		slog.Error("Failed to create default notification preferences for new local user", "error", err, "user_id", newUser.ID, "username", newUser.Username)
		// Don't fail the user creation, but log the error
	} else {
		slog.Info("Created default notification preferences for new local user", "user_id", newUser.ID, "username", newUser.Username)
	}

	// Return user info (without password hash)
	return ctx.JSON(fiber.Map{
		"id":         newUser.ID,
		"username":   newUser.Username,
		"email":      newUser.Email.String,
		"user_type":  "local",
		"created_at": newUser.CreatedAt,
	})
}
