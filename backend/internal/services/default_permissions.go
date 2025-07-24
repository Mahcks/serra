package services

import (
	"context"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/pkg/permissions"
	"github.com/mahcks/serra/pkg/structures"
)

// DefaultPermissionsService handles default permission assignment
type DefaultPermissionsService struct {
	db *repository.Queries
}

// NewDefaultPermissionsService creates a new default permissions service
func NewDefaultPermissionsService(db *repository.Queries) *DefaultPermissionsService {
	return &DefaultPermissionsService{
		db: db,
	}
}

// GetDefaultPermissions returns the list of permissions that should be assigned to new users
func (s *DefaultPermissionsService) GetDefaultPermissions(ctx context.Context) ([]string, error) {
	var defaultPerms []string

	// Map of permission to setting key
	permissionSettings := map[string]structures.Setting{
		// Owner permission (usually not recommended as default)
		permissions.Owner: structures.SettingDefaultOwner,
		// Admin permissions
		permissions.AdminUsers:    structures.SettingDefaultAdminUsers,
		permissions.AdminServices: structures.SettingDefaultAdminServices,
		permissions.AdminSystem:   structures.SettingDefaultAdminSystem,
		// Request permissions
		permissions.RequestMovies:                    structures.SettingDefaultRequestMovies,
		permissions.RequestSeries:                    structures.SettingDefaultRequestSeries,
		permissions.Request4KMovies:                  structures.SettingDefaultRequest4KMovies,
		permissions.Request4KSeries:                  structures.SettingDefaultRequest4KSeries,
		permissions.RequestAutoApproveMovies:         structures.SettingDefaultRequestAutoApproveMovies,
		permissions.RequestAutoApproveSeries:         structures.SettingDefaultRequestAutoApproveSeries,
		permissions.RequestAutoApprove4KMovies:       structures.SettingDefaultRequestAutoApprove4KMovies,
		permissions.RequestAutoApprove4KSeries:       structures.SettingDefaultRequestAutoApprove4KSeries,
		// Request management permissions
		permissions.RequestsView:    structures.SettingDefaultRequestsView,
		permissions.RequestsApprove: structures.SettingDefaultRequestsApprove,
		permissions.RequestsManage:  structures.SettingDefaultRequestsManage,
	}

	// Check each permission setting
	for permission, settingKey := range permissionSettings {
		setting, err := s.db.GetSetting(ctx, settingKey.String())
		if err != nil {
			// Setting doesn't exist, skip this permission
			continue
		}
		
		if setting == "true" {
			defaultPerms = append(defaultPerms, permission)
		}
	}

	return defaultPerms, nil
}

// AssignDefaultPermissions assigns default permissions to a user
func (s *DefaultPermissionsService) AssignDefaultPermissions(ctx context.Context, userID string) error {
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
			continue
		}
	}

	return nil
}