package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/internal/services/auth"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

// AuthenticateLocal handles authentication for both media server and local users
func (rg *RouteGroup) AuthenticateLocal(ctx *respond.Ctx) error {
	var req AuthRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Failed to parse request body")
	}

	// First, try local user authentication
	localUser, err := rg.gctx.Crate().Sqlite.Query().GetUserByUsername(ctx.Context(), req.Username)
	if err == nil {
		// Local user found, verify password
		if err := bcrypt.CompareHashAndPassword([]byte(localUser.PasswordHash.String), []byte(req.Password)); err != nil {
			return apiErrors.ErrUnauthorized().SetDetail("Invalid credentials")
		}

		// Create JWT token for local user
		token, _, err := rg.gctx.Crate().AuthService.CreateAccessToken(localUser.ID, localUser.Username, "", false) // Local users aren't admin by default
		if err != nil {
			return apiErrors.ErrInternalServerError().SetDetail("failed to create JWT token")
		}

		// Store the token in a cookie
		ctx.Cookie(rg.gctx.Crate().AuthService.Cookie(auth.CookieAuth, token, time.Hour*24*14))

		return ctx.SendStatus(fiber.StatusNoContent)
	}

	// If no local user found, try media server authentication (fallback to original flow)
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

	// Check if username already exists
	_, err = rg.gctx.Crate().Sqlite.Query().GetUserByUsername(ctx.Context(), req.Username)
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
		Username:     req.Username,
		Email:        utils.NewNullString(req.Email),
		PasswordHash: utils.NewNullString(string(hashedPassword)),
		AvatarUrl:    utils.NewNullString(""), // Local users start without avatars
	})
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to create user")
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
