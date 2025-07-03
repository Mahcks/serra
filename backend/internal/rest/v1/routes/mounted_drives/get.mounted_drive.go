package mounted_drives

import (
	"time"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

func (rg *RouteGroup) GetMountedDrive(ctx *respond.Ctx) error {
	/* user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	} */

	driveID := ctx.Params("id")
	if driveID == "" {
		return apiErrors.ErrBadRequest().SetDetail("drive ID is required")
	}

	drive, err := rg.gctx.Crate().Sqlite.Query().GetMountedDrive(ctx.Context(), driveID)
	if err != nil {
		return apiErrors.ErrNotFound().SetDetail("mounted drive not found")
	}

	return ctx.JSON(structures.MountedDrive{
		ID:              drive.ID,
		Name:            drive.Name,
		MountPath:       drive.MountPath,
		Filesystem:      utils.NullableString{NullString: drive.Filesystem}.ToPointer(),
		TotalSize:       utils.NullableInt64{NullInt64: drive.TotalSize}.ToPointer(),
		UsedSize:        utils.NullableInt64{NullInt64: drive.UsedSize}.ToPointer(),
		AvailableSize:   utils.NullableInt64{NullInt64: drive.AvailableSize}.ToPointer(),
		UsagePercentage: utils.NullableFloat64{NullFloat64: drive.UsagePercentage}.ToPointer(),
		IsOnline:        utils.NullableBool{NullBool: drive.IsOnline}.Or(false),
		LastChecked:     utils.NullableTime{NullTime: drive.LastChecked}.Or(time.Time{}),
		CreatedAt:       utils.NullableTime{NullTime: drive.CreatedAt}.Or(time.Time{}),
		UpdatedAt:       utils.NullableTime{NullTime: drive.UpdatedAt}.Or(time.Time{}),
	})
}
