package downloads

import (
	"github.com/mahcks/serra/config"
	"github.com/mahcks/serra/internal/global"
)

type RouteGroup struct {
	gctx global.Context
}

func NewRouteGroup(gctx global.Context) *RouteGroup {
	return &RouteGroup{
		gctx: gctx,
	}
}

func (rg *RouteGroup) Config() *config.Config {
	return rg.gctx.Crate().Config.Get()
}
