package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/pkg/permissions"
)

// DynamicDefaultPermissionsService handles default permission assignment using the new table-based approach
type DynamicDefaultPermissionsService struct {
	db *repository.Queries
}

// NewDynamicDefaultPermissionsService creates a new dynamic default permissions service
func NewDynamicDefaultPermissionsService(db *repository.Queries) *DynamicDefaultPermissionsService {
	return &DynamicDefaultPermissionsService{
		db: db,
	}
}

// EnsureAllPermissionsExist ensures all permissions from permissions.go exist in default_permissions table
func (s *DynamicDefaultPermissionsService) EnsureAllPermissionsExist(ctx context.Context) error {
	allPermissions := permissions.GetAllPermissions()

	for _, permission := range allPermissions {
		err := s.db.EnsureDefaultPermissionExists(ctx, permission)
		if err != nil {
			slog.Error("Failed to ensure default permission exists", "permission", permission, "error", err)
			// Continue with other permissions even if one fails
		}
	}

	return nil
}

// GetDefaultPermissions returns the list of permissions that should be assigned to new users
func (s *DynamicDefaultPermissionsService) GetDefaultPermissions(ctx context.Context) ([]string, error) {
	// Ensure all permissions exist in the table first
	if err := s.EnsureAllPermissionsExist(ctx); err != nil {
		slog.Warn("Failed to ensure all permissions exist", "error", err)
		// Continue anyway
	}

	enabledPerms, err := s.db.GetDefaultPermissions(ctx)
	if err != nil {
		return nil, err
	}

	var defaultPerms []string
	for _, perm := range enabledPerms {
		defaultPerms = append(defaultPerms, perm.PermissionID)
	}

	return defaultPerms, nil
}

// AssignDefaultPermissions assigns default permissions to a user
func (s *DynamicDefaultPermissionsService) AssignDefaultPermissions(ctx context.Context, userID string) error {
	defaultPerms, err := s.GetDefaultPermissions(ctx)
	if err != nil {
		return err
	}

	// Assign each default permission to the user
	for _, permission := range defaultPerms {
		err := s.db.AssignUserPermission(ctx, repository.AssignUserPermissionParams{
			UserID:       userID,
			PermissionID: permission,
		})
		if err != nil {
			// Log error but continue with other permissions
			// This handles cases where permission might already exist
			slog.Debug("Failed to assign permission (may already exist)", "user_id", userID, "permission", permission, "error", err)
			continue
		}
	}

	return nil
}

// GetAllDefaultPermissionSettings returns all permissions and their default status
func (s *DynamicDefaultPermissionsService) GetAllDefaultPermissionSettings(ctx context.Context) (map[string]bool, error) {
	// Ensure all permissions exist in the table first
	if err := s.EnsureAllPermissionsExist(ctx); err != nil {
		slog.Warn("Failed to ensure all permissions exist", "error", err)
	}

	settings, err := s.db.GetAllDefaultPermissionSettings(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool)
	for _, setting := range settings {
		result[setting.PermissionID] = setting.Enabled
	}

	return result, nil
}

// UpdateDefaultPermission updates whether a permission is enabled by default
func (s *DynamicDefaultPermissionsService) UpdateDefaultPermission(ctx context.Context, permissionID string, enabled bool) error {
	// Validate that this is a real permission
	if !permissions.IsValidPermission(permissionID) {
		return fmt.Errorf("invalid permission: %s", permissionID)
	}

	return s.db.UpdateDefaultPermission(ctx, repository.UpdateDefaultPermissionParams{
		PermissionID: permissionID,
		Enabled:      enabled,
	})
}
