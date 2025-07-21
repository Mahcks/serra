package analytics

import (
	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/internal/services/drive_monitor"
)

type RouteGroup struct {
	gctx         global.Context
	driveMonitor *drive_monitor.DriveMonitorService
}

func New(gctx global.Context) *RouteGroup {
	driveMonitorSvc := drive_monitor.NewDriveMonitorService(gctx.Crate().Sqlite.Query())
	
	return &RouteGroup{
		gctx:         gctx,
		driveMonitor: driveMonitorSvc,
	}
}