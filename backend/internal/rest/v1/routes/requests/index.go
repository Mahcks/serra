package requests

import (
	"context"

	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/internal/integrations"
	"github.com/mahcks/serra/internal/integrations/radarr"
	"github.com/mahcks/serra/internal/integrations/sonarr"
	"github.com/mahcks/serra/internal/services/request_processor"
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
	return rg.requestProcessor.ProcessApprovedRequest(ctx, requestID)
}