package mounted_drives

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) DeleteMountedDrive(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	driveID := ctx.Params("id")
	if driveID == "" {
		return apiErrors.ErrBadRequest().SetDetail("drive ID is required")
	}

	// Check if drive exists
	_, err := rg.gctx.Crate().Sqlite.Query().GetMountedDrive(ctx.Context(), driveID)
	if err != nil {
		return apiErrors.ErrNotFound().SetDetail("mounted drive not found")
	}

	// Delete drive
	err = rg.gctx.Crate().Sqlite.Query().DeleteMountedDrive(ctx.Context(), driveID)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to delete mounted drive")
	}

	return ctx.JSON(map[string]interface{}{
		"message": "Mounted drive deleted successfully",
		"id":      driveID,
	})
}
