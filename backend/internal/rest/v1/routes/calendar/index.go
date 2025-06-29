package calendar

import (
	"github.com/mahcks/serra/config"
	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/internal/integrations"
)

type RouteGroup struct {
	gctx         global.Context
	integrations *integrations.Integration
}

func NewRouteGroup(gctx global.Context, integrations *integrations.Integration) *RouteGroup {
	return &RouteGroup{
		gctx:         gctx,
		integrations: integrations,
	}
}

func (rg *RouteGroup) Config() *config.Config {
	return rg.gctx.Crate().Config.Get()
}
