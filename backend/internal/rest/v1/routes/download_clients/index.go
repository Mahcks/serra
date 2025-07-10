package downloadclients

import (
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
