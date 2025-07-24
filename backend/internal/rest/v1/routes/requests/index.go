package requests

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/internal/integrations"
	"github.com/mahcks/serra/internal/integrations/radarr"
	"github.com/mahcks/serra/internal/integrations/sonarr"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/internal/services/request_processor"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
)

type RouteGroup struct {
	gctx             global.Context
	requestProcessor request_processor.Service
}

func NewRouteGroup(gctx global.Context, integrations *integrations.Integration) *RouteGroup {
	// Initialize Radarr and Sonarr services
	radarrSvc := radarr.New(gctx.Crate().Sqlite.Query())
	sonarrSvc := sonarr.New(gctx.Crate().Sqlite.Query())

	// Initialize request processor
	processor := request_processor.New(gctx.Crate().Sqlite.Query(), radarrSvc, sonarrSvc, integrations)

	return &RouteGroup{
		gctx:             gctx,
		requestProcessor: processor,
	}
}

// processApprovedRequest handles the automatic processing of approved requests
func (rg *RouteGroup) processApprovedRequest(ctx context.Context, requestID int64) error {
	slog.Info("Processing approved request", "request_id", requestID)
	return rg.requestProcessor.ProcessApprovedRequest(ctx, requestID)
}

// RetryFailedRequests provides an endpoint to retry all failed requests
func (rg *RouteGroup) RetryFailedRequests(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Check if user has admin permissions
	if !user.IsAdmin {
		hasPermission, err := rg.checkUserPermissionForUpdate(ctx.Context(), user.ID, permissions.RequestsManage)
		if err != nil {
			return apiErrors.ErrInternalServerError().SetDetail("Permission check failed")
		}
		if !hasPermission {
			return apiErrors.ErrForbidden().SetDetail("You don't have permission to retry failed requests")
		}
	}

	err := rg.requestProcessor.RetryFailedRequests(ctx.Context())
	if err != nil {
		slog.Error("Failed to retry failed requests", "error", err)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to retry requests")
	}

	return ctx.JSON(fiber.Map{"message": "Failed requests retry initiated"})
}

// checkUserPermissionForUpdate checks if a user has a specific permission (shared utility)
func (rg *RouteGroup) checkUserPermissionForUpdate(ctx context.Context, userID, permission string) (bool, error) {
	userPermissions, err := rg.gctx.Crate().Sqlite.Query().GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	// Check if the user has owner permission (grants all access)
	for _, userPerm := range userPermissions {
		if userPerm.PermissionID == permissions.Owner {
			return true, nil
		}
	}

	// Check if the permission exists in the user's permissions
	for _, userPerm := range userPermissions {
		if userPerm.PermissionID == permission {
			return true, nil
		}
	}

	return false, nil
}
