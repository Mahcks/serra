package mounted_drives

import (
	"github.com/mahcks/serra/internal/jobs"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) GetSystemDrives(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Create a temporary drive monitor to get system drives
	job, err := jobs.NewJob("drive_monitor", rg.gctx)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to create drive monitor")
	}
	driveMonitor, ok := job.(*jobs.DriveMonitor)
	if !ok {
		return apiErrors.ErrInternalServerError().SetDetail("failed to cast to drive monitor")
	}

	drives, err := driveMonitor.GetSystemDrives()
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to get system drives")
	}

	return ctx.JSON(map[string]interface{}{
		"drives": drives,
	})
}
