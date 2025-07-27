package invitations

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/internal/services"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

func (rg *RouteGroup) GetInvitationByToken(ctx *respond.Ctx) error {
	token := ctx.Params("token")
	if token == "" {
		return apiErrors.ErrBadRequest().SetDetail("Token is required")
	}

	// Get invitation by token
	invitation, err := rg.gctx.Crate().Sqlite.Query().GetInvitationByToken(ctx.Context(), token)
	if err != nil {
		if err == sql.ErrNoRows {
			return apiErrors.ErrNotFound().SetDetail("Invalid or expired invitation")
		}
		utils.LogErrorWithStack("Failed to get invitation by token", err, "token", token)
		return apiErrors.ErrInternalServerError()
	}

	// Parse permissions
	var permissions []string
	if invitation.Permissions.Valid && invitation.Permissions.String != "" {
		if err := json.Unmarshal([]byte(invitation.Permissions.String), &permissions); err != nil {
			permissions = []string{}
		}
	}

	// Return invitation details (without sensitive info)
	apiInvitation := structures.Invitation{
		ID:              invitation.ID,
		Email:           invitation.Email,
		Username:        invitation.Username,
		// Don't include token in response
		Permissions:     permissions,
		CreateMediaUser: invitation.CreateMediaUser.Bool,
		Status:          invitation.Status.String,
		ExpiresAt:       invitation.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt:       invitation.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}

	return ctx.JSON(apiInvitation)
}

func (rg *RouteGroup) AcceptInvitation(ctx *respond.Ctx) error {
	var req structures.AcceptInvitationRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Invalid request body")
	}

	// Validate request
	if req.Token == "" || req.Password == "" {
		return apiErrors.ErrBadRequest().SetDetail("Token and password are required")
	}

	if req.Password != req.ConfirmPassword {
		return apiErrors.ErrBadRequest().SetDetail("Passwords do not match")
	}

	// Strengthen password requirements
	if err := validatePasswordStrength(req.Password); err != nil {
		return err
	}

	// Get invitation by token
	invitation, err := rg.gctx.Crate().Sqlite.Query().GetInvitationByToken(ctx.Context(), req.Token)
	if err != nil {
		if err == sql.ErrNoRows {
			return apiErrors.ErrNotFound().SetDetail("Invalid or expired invitation")
		}
		utils.LogErrorWithStack("Failed to get invitation by token", err, "token", req.Token)
		return apiErrors.ErrInternalServerError()
	}

	// Use a transaction to prevent race conditions during user creation
	tx, err := rg.gctx.Crate().Sqlite.DB().BeginTx(ctx.Context(), nil)
	if err != nil {
		utils.LogErrorWithStack("Failed to begin transaction", err)
		return apiErrors.ErrInternalServerError()
	}
	defer tx.Rollback()

	// Check if username is still available within the transaction
	existingUser, err := rg.gctx.Crate().Sqlite.Query().GetUserByUsername(ctx.Context(), strings.ToLower(invitation.Username))
	if err == nil && existingUser.ID != "" {
		return apiErrors.ErrConflict().SetDetail("Username is no longer available")
	}

	// Note: We no longer use invitation permissions - instead we'll assign default permissions

	var newUser repository.User
	var mediaServerUserID string

	if invitation.CreateMediaUser.Bool {
		// Create media server user first
		mediaUserID, err := rg.createMediaServerUser(ctx.Context(), invitation.Username, invitation.Email, req.Password)
		if err != nil {
			// Fail the invitation if media server creation fails when it was requested
			utils.LogErrorWithStack("Failed to create media server user for invitation", err,
				"username", invitation.Username,
				"email", invitation.Email)
			return apiErrors.ErrInternalServerError().SetDetail("Failed to create media server account")
		}
		mediaServerUserID = mediaUserID

		// Create Serra media account using the media server user ID
		avatarURL := fmt.Sprintf("/users/%s/avatar", mediaUserID)
		newUser, err = rg.gctx.Crate().Sqlite.Query().CreateUser(ctx.Context(), repository.CreateUserParams{
			ID:           mediaUserID, // Use media server user ID
			Username:     strings.ToLower(invitation.Username),
			AccessToken:  sql.NullString{Valid: false}, // No access token for invitation-created users
			Email:        sql.NullString{String: invitation.Email, Valid: invitation.Email != ""},
			AvatarUrl:    sql.NullString{String: avatarURL, Valid: true},
			UserType:     "media_server",
			PasswordHash: sql.NullString{Valid: false}, // Media server users don't store password hash in Serra
		})
		if err != nil {
			// Check if it's a unique constraint violation (race condition)
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				return apiErrors.ErrConflict().SetDetail("Username is no longer available")
			}
			utils.LogErrorWithStack("Failed to create media user from invitation", err,
				"username", invitation.Username,
				"email", invitation.Email,
				"invitation_id", invitation.ID,
				"media_user_id", mediaUserID)
			return apiErrors.ErrInternalServerError()
		}

		slog.Info("Created media server account from invitation",
			"user_id", newUser.ID,
			"username", newUser.Username,
			"media_user_id", mediaUserID)
	} else {
		// Create local user
		userID, err := generateUserID()
		if err != nil {
			utils.LogErrorWithStack("Failed to generate user ID", err)
			return apiErrors.ErrInternalServerError()
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.LogErrorWithStack("Failed to hash password", err)
			return apiErrors.ErrInternalServerError().SetDetail("Failed to hash password")
		}

		// Create Serra local user
		newUser, err = rg.gctx.Crate().Sqlite.Query().CreateLocalUser(ctx.Context(), repository.CreateLocalUserParams{
			ID:                   userID,
			Username:             strings.ToLower(invitation.Username),
			Email:                sql.NullString{String: invitation.Email, Valid: invitation.Email != ""},
			PasswordHash:         sql.NullString{String: string(hashedPassword), Valid: true},
			AvatarUrl:            sql.NullString{Valid: false},
			InvitedBy:            sql.NullString{String: invitation.InvitedBy, Valid: true},
			InvitationAcceptedAt: sql.NullTime{Time: time.Now(), Valid: true},
		})
		if err != nil {
			// Check if it's a unique constraint violation (race condition)
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				return apiErrors.ErrConflict().SetDetail("Username is no longer available")
			}
			utils.LogErrorWithStack("Failed to create local user from invitation", err,
				"username", invitation.Username,
				"email", invitation.Email,
				"invitation_id", invitation.ID)
			return apiErrors.ErrInternalServerError()
		}

		slog.Info("Created local account from invitation",
			"user_id", newUser.ID,
			"username", newUser.Username)
	}

	// Assign default permissions to the new user (instead of invitation-specific permissions)
	defaultPermissionsService := services.NewDynamicDefaultPermissionsService(rg.gctx.Crate().Sqlite.Query())
	if err := defaultPermissionsService.AssignDefaultPermissions(ctx.Context(), newUser.ID); err != nil {
		slog.Error("Failed to assign default permissions to new user from invitation",
			"error", err,
			"user_id", newUser.ID,
			"username", newUser.Username)
		// Don't fail the invitation - user can be granted permissions later
	} else {
		slog.Info("Assigned default permissions to new user from invitation",
			"user_id", newUser.ID,
			"username", newUser.Username)
	}

	// Mark invitation as accepted
	_, err = rg.gctx.Crate().Sqlite.Query().UpdateInvitationStatus(ctx.Context(), repository.UpdateInvitationStatusParams{
		Status: sql.NullString{String: "accepted", Valid: true},
		Token:  req.Token,
	})
	if err != nil {
		utils.LogErrorWithStack("Failed to update invitation status", err,
			"token", req.Token,
			"invitation_id", invitation.ID)
		// Don't fail - user is created successfully
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		utils.LogErrorWithStack("Failed to commit transaction", err)
		return apiErrors.ErrInternalServerError()
	}

	// Audit log for invitation acceptance
	slog.Info("AUDIT: Invitation accepted",
		"action", "invitation.accept",
		"invitation_id", invitation.ID,
		"user_id", newUser.ID,
		"username", invitation.Username,
		"email", invitation.Email,
		"invited_by", invitation.InvitedBy,
		"user_type", newUser.UserType,
		"media_server_user_created", invitation.CreateMediaUser.Bool && mediaServerUserID != "",
		"media_server_user_id", mediaServerUserID,
		"client_ip", ctx.IP(),
		"user_agent", ctx.Get("User-Agent"))

	// Return success response
	return ctx.Status(201).JSON(map[string]interface{}{
		"message": "Invitation accepted successfully",
		"user": map[string]interface{}{
			"id":         newUser.ID,
			"username":   newUser.Username,
			"email":      newUser.Email.String,
			"user_type":  newUser.UserType,
		},
		"media_server_account_created": invitation.CreateMediaUser.Bool && mediaServerUserID != "",
	})
}

func (rg *RouteGroup) createMediaServerUser(ctx context.Context, username, _ string, password string) (string, error) {
	// Get media server type from settings to determine which integration to use
	cfg := rg.gctx.Crate().Config.Get()
	
	switch cfg.MediaServer.Type {
	case "emby", "jellyfin":
		// Both Emby and Jellyfin use the same API structure for user creation
		return rg.integrations.Emby.CreateUser(ctx, username, password)
	default:
		slog.Warn("Unsupported media server type for user creation",
			"type", cfg.MediaServer.Type,
			"username", username)
		return "", fmt.Errorf("unsupported media server type: %s", cfg.MediaServer.Type)
	}
}


func generateUserID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// validatePasswordStrength validates password strength requirements
func validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return apiErrors.ErrBadRequest().SetDetail("Password must be at least 8 characters long")
	}
	
	if len(password) > 128 {
		return apiErrors.ErrBadRequest().SetDetail("Password must be no more than 128 characters long")
	}
	
	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	
	if !hasUpper {
		return apiErrors.ErrBadRequest().SetDetail("Password must contain at least one uppercase letter")
	}
	
	if !hasLower {
		return apiErrors.ErrBadRequest().SetDetail("Password must contain at least one lowercase letter")
	}
	
	if !hasNumber {
		return apiErrors.ErrBadRequest().SetDetail("Password must contain at least one number")
	}
	
	if !hasSpecial {
		return apiErrors.ErrBadRequest().SetDetail("Password must contain at least one special character")
	}
	
	// Check for common weak patterns
	lowercasePassword := strings.ToLower(password)
	weakPatterns := []string{
		"password", "123456", "qwerty", "admin", "letmein",
		"welcome", "monkey", "dragon", "master", "login",
	}
	
	for _, pattern := range weakPatterns {
		if strings.Contains(lowercasePassword, pattern) {
			return apiErrors.ErrBadRequest().SetDetail("Password contains common weak patterns")
		}
	}
	
	// Check for repetitive characters (e.g., "aaaa", "1111")
	repetitivePattern := regexp.MustCompile(`(.)\1{3,}`)
	if repetitivePattern.MatchString(password) {
		return apiErrors.ErrBadRequest().SetDetail("Password cannot contain repetitive characters")
	}
	
	// Check for sequential patterns (e.g., "1234", "abcd")
	if containsSequentialPattern(password) {
		return apiErrors.ErrBadRequest().SetDetail("Password cannot contain sequential patterns")
	}
	
	return nil
}

// containsSequentialPattern checks for sequential character patterns
func containsSequentialPattern(password string) bool {
	// Check for ascending sequences (at least 4 chars)
	for i := 0; i <= len(password)-4; i++ {
		if isSequential(password[i:i+4], true) {
			return true
		}
	}
	
	// Check for descending sequences (at least 4 chars)
	for i := 0; i <= len(password)-4; i++ {
		if isSequential(password[i:i+4], false) {
			return true
		}
	}
	
	return false
}

// isSequential checks if a string contains sequential characters
func isSequential(s string, ascending bool) bool {
	for i := 1; i < len(s); i++ {
		if ascending {
			if s[i] != s[i-1]+1 {
				return false
			}
		} else {
			if s[i] != s[i-1]-1 {
				return false
			}
		}
	}
	return true
}