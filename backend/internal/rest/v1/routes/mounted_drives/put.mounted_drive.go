package mounted_drives

import (
	"time"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

func (rg *RouteGroup) UpdateMountedDrive(ctx *respond.Ctx) error {
	/* user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	} */

	driveID := ctx.Params("id")
	if driveID == "" {
		return apiErrors.ErrBadRequest().SetDetail("drive ID is required")
	}

	var req structures.UpdateMountedDriveRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("invalid request body")
	}

	// Validate required fields
	if req.Name == "" || req.MountPath == "" {
		return apiErrors.ErrBadRequest().SetDetail("name and mount_path are required")
	}

	// Check if drive exists
	existing, err := rg.gctx.Crate().Sqlite.Query().GetMountedDrive(ctx.Context(), driveID)
	if err != nil {
		return apiErrors.ErrNotFound().SetDetail("mounted drive not found")
	}

	// Check if mount path already exists (for different drive)
	if req.MountPath != existing.MountPath {
		pathConflict, err := rg.gctx.Crate().Sqlite.Query().GetMountedDriveByPath(ctx.Context(), req.MountPath)
		if err == nil && pathConflict.ID != "" && pathConflict.ID != driveID {
			return apiErrors.ErrConflict().SetDetail("mount path already exists")
		}
	}

	// Update drive
	params := repository.UpdateMountedDriveParams{
		Name:            req.Name,
		MountPath:       req.MountPath,
		Filesystem:      existing.Filesystem,
		TotalSize:       existing.TotalSize,
		UsedSize:        existing.UsedSize,
		AvailableSize:   existing.AvailableSize,
		UsagePercentage: existing.UsagePercentage,
		IsOnline:        existing.IsOnline,
		LastChecked:     existing.LastChecked,
		ID:              driveID,
	}

	err = rg.gctx.Crate().Sqlite.Query().UpdateMountedDrive(ctx.Context(), params)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to update mounted drive")
	}

	return ctx.JSON(map[string]interface{}{
		"id":         driveID,
		"name":       req.Name,
		"mount_path": req.MountPath,
		"updated_at": time.Now(),
	})
}
