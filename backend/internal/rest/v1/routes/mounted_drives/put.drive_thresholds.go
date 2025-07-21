package mounted_drives

import (
	"database/sql"
	"log/slog"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

func (rg *RouteGroup) PutDriveThresholds(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	driveID := ctx.Params("id")
	if driveID == "" {
		return apiErrors.ErrBadRequest().SetDetail("Drive ID is required")
	}

	var req structures.UpdateDriveThresholdsRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Invalid request body")
	}

	// Validate thresholds if provided
	if req.WarningThreshold != nil && (*req.WarningThreshold < 0 || *req.WarningThreshold > 100) {
		return apiErrors.ErrBadRequest().SetDetail("Warning threshold must be between 0 and 100")
	}
	if req.CriticalThreshold != nil && (*req.CriticalThreshold < 0 || *req.CriticalThreshold > 100) {
		return apiErrors.ErrBadRequest().SetDetail("Critical threshold must be between 0 and 100")
	}
	if req.GrowthRateThreshold != nil && *req.GrowthRateThreshold < 0 {
		return apiErrors.ErrBadRequest().SetDetail("Growth rate threshold must be non-negative")
	}

	// Check if drive exists
	_, err := rg.gctx.Crate().Sqlite.Query().GetMountedDrive(ctx.Context(), driveID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apiErrors.ErrNotFound().SetDetail("Drive not found")
		}
		return apiErrors.ErrInternalServerError().SetDetail("Failed to get drive")
	}

	// Update thresholds
	err = rg.gctx.Crate().Sqlite.Query().UpdateDriveThresholds(ctx.Context(), repository.UpdateDriveThresholdsParams{
		WarningThreshold:    utils.NewNullFloat64FromPtr(req.WarningThreshold),
		CriticalThreshold:   utils.NewNullFloat64FromPtr(req.CriticalThreshold),
		GrowthRateThreshold: utils.NewNullFloat64FromPtr(req.GrowthRateThreshold),
		MonitoringEnabled:   utils.NewNullBoolFromPtr(req.MonitoringEnabled),
		ID:                  driveID,
	})
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to update drive thresholds")
	}

	// Clear existing alerts for this drive when:
	// 1. Thresholds have changed (they may no longer be relevant)
	// 2. Monitoring has been disabled (alerts should not remain active)
	// This prevents stale alerts and ensures fresh evaluation with new settings
	shouldClearAlerts := req.WarningThreshold != nil ||
		req.CriticalThreshold != nil ||
		req.GrowthRateThreshold != nil ||
		(req.MonitoringEnabled != nil && !*req.MonitoringEnabled)

	if shouldClearAlerts {
		err = rg.gctx.Crate().Sqlite.Query().ClearDriveAlerts(ctx.Context(), driveID)
		if err != nil {
			// Log error but don't fail the request - threshold update was successful
			// The drive monitoring will recreate appropriate alerts on the next run
			slog.Error("Failed to clear old drive alerts after threshold update",
				"drive_id", driveID, "error", err)
		}
	}

	return ctx.JSON(map[string]string{"message": "Drive thresholds updated successfully"})
}
