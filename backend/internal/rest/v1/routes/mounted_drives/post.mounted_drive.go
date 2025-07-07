package mounted_drives

import (
	"time"

	"github.com/google/uuid"
	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

func (rg *RouteGroup) CreateMountedDrive(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	var req structures.CreateMountedDriveRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("invalid request body")
	}

	// Validate required fields
	if req.Name == "" || req.MountPath == "" {
		return apiErrors.ErrBadRequest().SetDetail("name and mount_path are required")
	}

	// Check if mount path already exists
	existing, err := rg.gctx.Crate().Sqlite.Query().GetMountedDriveByPath(ctx.Context(), req.MountPath)
	if err == nil && existing.ID != "" {
		return apiErrors.ErrConflict().SetDetail("mount path already exists")
	}

	// Create new drive
	driveID := uuid.New().String()
	now := time.Now()

	params := repository.CreateMountedDriveParams{
		ID:              driveID,
		Name:            req.Name,
		MountPath:       req.MountPath,
		Filesystem:      utils.NewNullString(""),
		TotalSize:       utils.NewNullInt64(0, false),
		UsedSize:        utils.NewNullInt64(0, false),
		AvailableSize:   utils.NewNullInt64(0, false),
		UsagePercentage: utils.NewNullFloat64(0, false),
		IsOnline:        utils.NewNullBool(true),
		LastChecked:     utils.NewNullTime(now),
	}

	err = rg.gctx.Crate().Sqlite.Query().CreateMountedDrive(ctx.Context(), params)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to create mounted drive")
	}

	return ctx.JSON(map[string]interface{}{
		"id":         driveID,
		"name":       req.Name,
		"mount_path": req.MountPath,
		"created_at": now,
	})
}
