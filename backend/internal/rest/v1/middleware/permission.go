package middleware

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
)

// RequirePermission creates middleware that checks if the user has the required permission
func RequirePermission(db *repository.Queries, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := &respond.Ctx{Ctx: c}

		// Get user from JWT context (set by JWT middleware)
		userClaims := ctx.ParseClaims()
		if userClaims == nil {
			slog.Error("Failed to parse JWT claims")
			return apiErrors.ErrBadRequest().SetDetail("Invalid authentication token")
		}

		slog.Info("Permission check", "user_id", userClaims.ID, "username", userClaims.Username, "is_admin", userClaims.IsAdmin, "required_permission", permission)

		// Check if user is admin (admins have all permissions)
		if userClaims.IsAdmin {
			slog.Info("Admin user accessing endpoint", "user_id", userClaims.ID, "username", userClaims.Username)
			return c.Next()
		}

		// Check if user has the specific permission or owner permission
		hasPermission, err := checkUserPermission(c.Context(), db, userClaims.ID, permission)
		if err != nil {
			slog.Error("Failed to check user permission", "error", err, "user_id", userClaims.ID, "permission", permission)
			return apiErrors.ErrInternalServerError().SetDetail("Permission check failed")
		}

		slog.Info("Permission check result", "user_id", userClaims.ID, "username", userClaims.Username, "permission", permission, "has_permission", hasPermission)

		if !hasPermission {
			slog.Warn("User lacks required permission", "user_id", userClaims.ID, "username", userClaims.Username, "permission", permission)
			return apiErrors.ErrForbidden().SetDetail(fmt.Sprintf("Missing required permission: %s", permission))
		}

		return c.Next()
	}
}

// RequireAnyPermission creates middleware that checks if the user has any of the specified permissions
func RequireAnyPermission(db *repository.Queries, permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := &respond.Ctx{Ctx: c}

		// Get user from JWT context
		userClaims := ctx.ParseClaims()
		if userClaims == nil {
			slog.Error("Failed to parse JWT claims")
			return apiErrors.ErrBadRequest().SetDetail("Invalid authentication token")
		}

		// Admins have all permissions
		if userClaims.IsAdmin {
			return c.Next()
		}

		// Check if user has any of the required permissions or owner permission
		for _, permission := range permissions {
			hasPermission, err := checkUserPermission(c.Context(), db, userClaims.ID, permission)
			if err != nil {
				slog.Error("Failed to check user permission", "error", err, "user_id", userClaims.ID, "permission", permission)
				continue
			}

			if hasPermission {
				return c.Next()
			}
		}

		slog.Warn("User lacks any required permission", "user_id", userClaims.ID, "username", userClaims.Username, "permissions", permissions)
		return apiErrors.ErrForbidden().SetDetail(fmt.Sprintf("Missing required permissions: %v", permissions))
	}
}

// RequireAllPermissions creates middleware that checks if the user has all of the specified permissions
func RequireAllPermissions(db *repository.Queries, permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := &respond.Ctx{Ctx: c}

		// Get user from JWT context
		userClaims := ctx.ParseClaims()
		if userClaims == nil {
			slog.Error("Failed to parse JWT claims")
			return apiErrors.ErrBadRequest().SetDetail("Invalid authentication token")
		}

		// Admins have all permissions
		if userClaims.IsAdmin {
			return c.Next()
		}

		// Check if user has all required permissions or owner permission
		for _, permission := range permissions {
			hasPermission, err := checkUserPermission(c.Context(), db, userClaims.ID, permission)
			if err != nil {
				slog.Error("Failed to check user permission", "error", err, "user_id", userClaims.ID, "permission", permission)
				return apiErrors.ErrInternalServerError().SetDetail("Permission check failed")
			}

			if !hasPermission {
				slog.Warn("User lacks required permission", "user_id", userClaims.ID, "username", userClaims.Username, "permission", permission)
				return apiErrors.ErrForbidden().SetDetail(fmt.Sprintf("Missing required permission: %s", permission))
			}
		}

		return c.Next()
	}
}

// RequireAdmin creates middleware that checks if the user is an admin
func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := &respond.Ctx{Ctx: c}

		// Get user from JWT context
		userClaims := ctx.ParseClaims()
		if userClaims == nil {
			slog.Error("Failed to parse JWT claims")
			return apiErrors.ErrBadRequest().SetDetail("Invalid authentication token")
		}

		if !userClaims.IsAdmin {
			slog.Warn("Non-admin user attempted to access admin endpoint", "user_id", userClaims.ID, "username", userClaims.Username)
			return apiErrors.ErrForbidden().SetDetail("Admin access required")
		}

		return c.Next()
	}
}

// checkUserPermission checks if a user has a specific permission or owner permission
func checkUserPermission(ctx context.Context, db *repository.Queries, userID, permission string) (bool, error) {
	// Query user_permissions table to check if user has the permission
	userPermissions, err := db.GetUserPermissions(ctx, userID)
	if err != nil {
		slog.Error("Database query failed", "error", err, "user_id", userID)
		return false, fmt.Errorf("failed to get user permissions: %w", err)
	}

	slog.Info("User permissions from database", "user_id", userID, "permissions_count", len(userPermissions))
	for i, userPerm := range userPermissions {
		slog.Info("User permission", "index", i, "permission_id", userPerm.PermissionID, "user_id", userID)
	}

	// Check if the user has owner permission (grants all access)
	for _, userPerm := range userPermissions {
		if userPerm.PermissionID == permissions.Owner {
			slog.Info("Owner permission found - granting access", "user_id", userID, "requested_permission", permission)
			return true, nil
		}
	}

	// Check if the permission exists in the user's permissions
	for _, userPerm := range userPermissions {
		if userPerm.PermissionID == permission {
			slog.Info("Permission match found", "user_id", userID, "permission", permission)
			return true, nil
		}
	}

	slog.Warn("No permission match found", "user_id", userID, "required_permission", permission)
	return false, nil
}
