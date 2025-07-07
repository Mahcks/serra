package mounted_drives

import (
	"time"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

func (rg *RouteGroup) GetMountedDrives(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	drives, err := rg.gctx.Crate().Sqlite.Query().ListMountedDrives(ctx.Context())
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to fetch mounted drives")
	}

	if len(drives) == 0 {
		return ctx.JSON([]structures.MountedDrive{})
	}

	// Convert to response format
	var response []structures.MountedDrive
	for _, drive := range drives {
		mountedDrive := structures.MountedDrive{
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
		}
		response = append(response, mountedDrive)
	}

	return ctx.JSON(response)
}
