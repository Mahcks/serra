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
	"strconv"
	"strings"
	"time"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/internal/services/email"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

func (rg *RouteGroup) CreateInvitation(ctx *respond.Ctx) error {
	// Parse claims to get the inviting user
	user := ctx.ParseClaims()
	if user == nil {
		return apiErrors.ErrUnauthorized()
	}

	// Check if user has permission to invite others
	hasPermission, err := rg.checkInvitePermission(ctx.Context(), user.ID)
	if err != nil {
		utils.LogErrorWithStack("Failed to check invite permission", err, "user_id", user.ID)
		return apiErrors.ErrInternalServerError()
	}
	if !hasPermission {
		return apiErrors.ErrForbidden().SetDetail("You don't have permission to invite users")
	}

	// Parse request body
	var req structures.CreateInvitationRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Invalid request body")
	}

	// Validate request
	if req.Email == "" || req.Username == "" {
		return apiErrors.ErrBadRequest().SetDetail("Email and username are required")
	}

	// Validate email format
	if !isValidEmail(req.Email) {
		return apiErrors.ErrBadRequest().SetDetail("Invalid email format")
	}

	// Validate username format (alphanumeric, dots, underscores, hyphens)
	if !isValidUsername(req.Username) {
		return apiErrors.ErrBadRequest().SetDetail("Username must contain only letters, numbers, dots, underscores, and hyphens")
	}

	// Validate permissions array
	if err := rg.validatePermissions(ctx.Context(), req.Permissions); err != nil {
		return err
	}

	// Check if user already exists
	existingUser, err := rg.gctx.Crate().Sqlite.Query().GetUserByUsername(ctx.Context(), req.Username)
	if err == nil && existingUser.ID != "" {
		return apiErrors.ErrConflict().SetDetail("Username already exists")
	}

	// Check if there's already a pending invitation for this email
	existingInvite, err := rg.gctx.Crate().Sqlite.Query().GetInvitationByEmail(ctx.Context(), req.Email)
	if err == nil && existingInvite.Status.String == "pending" {
		return apiErrors.ErrConflict().SetDetail("Pending invitation already exists for this email")
	}

	// Set default expiration
	expiresInDays := req.ExpiresInDays
	if expiresInDays <= 0 {
		expiresInDays = 7 // Default to 7 days
	}
	expiresAt := time.Now().AddDate(0, 0, expiresInDays)

	// Generate secure token
	token, err := generateInviteToken()
	if err != nil {
		utils.LogErrorWithStack("Failed to generate invite token", err)
		return apiErrors.ErrInternalServerError()
	}

	// Serialize permissions
	permissionsJSON := "[]"
	if len(req.Permissions) > 0 {
		permData, err := json.Marshal(req.Permissions)
		if err != nil {
			utils.LogErrorWithStack("Failed to marshal permissions", err)
			return apiErrors.ErrInternalServerError()
		}
		permissionsJSON = string(permData)
	}

	// Create invitation in database
	invitation, err := rg.gctx.Crate().Sqlite.Query().CreateInvitation(ctx.Context(), repository.CreateInvitationParams{
		Email:           req.Email,
		Username:        req.Username,
		Token:           token,
		InvitedBy:       user.ID,
		Permissions:     sql.NullString{String: permissionsJSON, Valid: true},
		CreateMediaUser: sql.NullBool{Bool: req.CreateMediaUser, Valid: true},
		ExpiresAt:       expiresAt,
	})

	if err != nil {
		utils.LogErrorWithStack("Failed to create invitation", err,
			"email", req.Email,
			"username", req.Username,
			"invited_by", user.ID)
		return apiErrors.ErrInternalServerError()
	}

	// Get inviter details for email
	inviter, err := rg.gctx.Crate().Sqlite.Query().GetUserByID(ctx.Context(), user.ID)
	if err != nil {
		utils.LogErrorWithStack("Failed to get inviter details", err, "user_id", user.ID)
		return apiErrors.ErrInternalServerError()
	}

	// Send invitation email (if email service is enabled)
	// This is optional - admins can copy and share links manually
	emailInvitation := structures.Invitation{
		ID:              invitation.ID,
		Email:           invitation.Email,
		Username:        invitation.Username,
		Token:           invitation.Token,
		InvitedBy:       invitation.InvitedBy,
		CreateMediaUser: invitation.CreateMediaUser.Bool,
		Status:          invitation.Status.String,
		ExpiresAt:       invitation.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt:       invitation.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       invitation.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	go rg.sendInvitationEmail(emailInvitation, inviter.Username)

	// Convert to API response format (include token for immediate sharing)
	appURL := rg.getSettingWithDefault("app_url", "http://localhost:3000")
	inviteURL := fmt.Sprintf("%s/invite/accept/%s", appURL, invitation.Token)
	
	apiInvitation := structures.Invitation{
		ID:              invitation.ID,
		Email:           invitation.Email,
		Username:        invitation.Username,
		Token:           invitation.Token, // Include token for immediate copying
		InvitedBy:       invitation.InvitedBy,
		Permissions:     req.Permissions,
		CreateMediaUser: invitation.CreateMediaUser.Bool,
		Status:          invitation.Status.String,
		ExpiresAt:       invitation.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt:       invitation.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       invitation.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}

	// Audit log for invitation creation
	slog.Info("AUDIT: Invitation created",
		"action", "invitation.create",
		"invitation_id", invitation.ID,
		"email", req.Email,
		"username", req.Username,
		"invited_by", user.ID,
		"invited_by_username", user.Username,
		"permissions", req.Permissions,
		"create_media_user", req.CreateMediaUser,
		"expires_at", expiresAt,
		"client_ip", ctx.IP(),
		"user_agent", ctx.Get("User-Agent"))

	return ctx.Status(201).JSON(map[string]interface{}{
		"invitation": apiInvitation,
		"invite_url": inviteURL,
	})
}

func (rg *RouteGroup) sendInvitationEmail(invitation structures.Invitation, inviterName string) {
	// Get email settings from database
	emailSettings, err := rg.getEmailSettingsFromDB(context.Background())
	if err != nil {
		slog.Error("Failed to retrieve email settings",
			"error", err,
			"invitation_id", invitation.ID)
		return
	}

	emailService := email.NewService(emailSettings)

	if !emailService.IsEnabled() {
		slog.Info("Email service not enabled, skipping invitation email",
			"invitation_id", invitation.ID)
		return
	}

	// Get app settings for URLs and names
	appName := rg.getSettingWithDefault("app_name", "Serra")
	appURL := rg.getSettingWithDefault("app_url", "http://localhost:3000")
	mediaServerName := rg.getSettingWithDefault("media_server_name", "Media Server")
	mediaServerURL := rg.getSettingWithDefault("media_server_url", "http://localhost:8096")

	// Prepare email data
	data := structures.InvitationEmailData{
		Username:        invitation.Username,
		InviterName:     inviterName,
		AppName:         appName,
		AppURL:          appURL,
		AcceptURL:       fmt.Sprintf("%s/invite/accept/%s", appURL, invitation.Token),
		MediaServerName: mediaServerName,
		MediaServerURL:  mediaServerURL,
		ExpiresAt:       invitation.ExpiresAt,
	}

	// Send invitation email in the foreground (could be moved to background job)
	if err := emailService.SendInvitation(data); err != nil {
		slog.Error("Failed to send invitation email",
			"error", err,
			"invitation_id", invitation.ID,
			"email", invitation.Email)
	} else {
		slog.Info("Invitation email sent successfully",
			"invitation_id", invitation.ID,
			"email", invitation.Email)
	}
}

// getEmailSettingsFromDB retrieves email settings from the database and constructs EmailSettings struct
func (rg *RouteGroup) getEmailSettingsFromDB(ctx context.Context) (*structures.EmailSettings, error) {
	getBoolSetting := func(key string, defaultVal bool) bool {
		if val, err := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx, key); err == nil {
			return val == "true"
		}
		return defaultVal
	}

	getStringSetting := func(key string, defaultVal string) string {
		if val, err := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx, key); err == nil {
			return val
		}
		return defaultVal
	}

	getIntSetting := func(key string, defaultVal int) int {
		if val, err := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx, key); err == nil {
			if intVal, err := strconv.Atoi(val); err == nil {
				return intVal
			}
		}
		return defaultVal
	}

	return &structures.EmailSettings{
		Enabled:            getBoolSetting("email_enabled", false),
		SenderName:         getStringSetting("email_sender_name", "Serra"),
		SenderAddress:      getStringSetting("email_sender_address", ""),
		SMTPHost:           getStringSetting("email_smtp_host", ""),
		SMTPPort:           getIntSetting("email_smtp_port", 587),
		SMTPUsername:       getStringSetting("email_smtp_username", ""),
		SMTPPassword:       getStringSetting("email_smtp_password", ""),
		EncryptionMethod:   getStringSetting("email_encryption_method", "starttls"),
		UseSTARTTLS:        getBoolSetting("email_use_starttls", true),
		AllowSelfSigned:    getBoolSetting("email_allow_self_signed", false),
	}, nil
}

// getSettingWithDefault retrieves a setting with a fallback default value
func (rg *RouteGroup) getSettingWithDefault(key, defaultVal string) string {
	if val, err := rg.gctx.Crate().Sqlite.Query().GetSetting(context.Background(), key); err == nil {
		return val
	}
	return defaultVal
}

func generateInviteToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (rg *RouteGroup) checkInvitePermission(ctx context.Context, userID string) (bool, error) {
	// Check if user has owner permission
	hasOwnerPermission, err := rg.gctx.Crate().Sqlite.Query().CheckUserPermission(ctx, repository.CheckUserPermissionParams{
		UserID:       userID,
		PermissionID: "owner",
	})
	if err == nil && hasOwnerPermission {
		return true, nil
	}

	// Check if user has admin.users permission
	hasAdminUsersPermission, err := rg.gctx.Crate().Sqlite.Query().CheckUserPermission(ctx, repository.CheckUserPermissionParams{
		UserID:       userID,
		PermissionID: "admin.users",
	})
	if err == nil && hasAdminUsersPermission {
		return true, nil
	}

	return false, nil
}

// isValidEmail validates email format using a comprehensive regex
func isValidEmail(email string) bool {
	// Comprehensive email regex that follows RFC 5322 standards
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	
	// Additional checks for security
	email = strings.TrimSpace(email)
	if len(email) > 254 { // RFC 5321 limit
		return false
	}
	
	// Check for common malicious patterns
	if strings.Contains(email, "..") || strings.HasPrefix(email, ".") || strings.HasSuffix(email, ".") {
		return false
	}
	
	return re.MatchString(email)
}

// isValidUsername validates username format
func isValidUsername(username string) bool {
	// Allow alphanumeric characters, dots, underscores, and hyphens
	// Must start with alphanumeric, length 3-50 characters
	usernameRegex := `^[a-zA-Z0-9][a-zA-Z0-9._-]{2,49}$`
	re := regexp.MustCompile(usernameRegex)
	
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 50 {
		return false
	}
	
	// Prevent consecutive special characters
	if strings.Contains(username, "..") || strings.Contains(username, "__") || strings.Contains(username, "--") {
		return false
	}
	
	return re.MatchString(username)
}

// validatePermissions validates the permissions array to ensure all permissions exist
func (rg *RouteGroup) validatePermissions(ctx context.Context, permissions []string) error {
	if len(permissions) == 0 {
		return nil // Empty permissions array is valid
	}
	
	// Limit the number of permissions to prevent abuse
	if len(permissions) > 50 {
		return apiErrors.ErrBadRequest().SetDetail("Too many permissions specified (max 50)")
	}
	
	// Check each permission exists in the database
	for _, permissionID := range permissions {
		// Validate permission ID format (prevent injection)
		if !isValidPermissionID(permissionID) {
			return apiErrors.ErrBadRequest().SetDetail(fmt.Sprintf("Invalid permission ID format: %s", permissionID))
		}
		
		// Check if permission exists
		_, err := rg.gctx.Crate().Sqlite.Query().GetPermission(ctx, permissionID)
		if err != nil {
			return apiErrors.ErrBadRequest().SetDetail(fmt.Sprintf("Permission not found: %s", permissionID))
		}
	}
	
	return nil
}

// isValidPermissionID validates permission ID format
func isValidPermissionID(permissionID string) bool {
	// Permission IDs should follow the format: category.action or just category
	// Examples: "admin.users", "requests.view", "owner"
	permissionRegex := `^[a-zA-Z][a-zA-Z0-9]*(\.[a-zA-Z][a-zA-Z0-9]*)?$`
	re := regexp.MustCompile(permissionRegex)
	
	if len(permissionID) > 100 { // Reasonable limit
		return false
	}
	
	return re.MatchString(permissionID)
}