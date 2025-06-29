package integrations

import (
	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/internal/integrations/jellystat"
	"github.com/mahcks/serra/internal/integrations/radarr"
	"github.com/mahcks/serra/internal/integrations/sonarr"
)

type Integration struct {
	Radarr    radarr.Service
	Sonarr    sonarr.Service
	Jellystat jellystat.Service
}

func New(gctx global.Context) *Integration {
	return &Integration{
		Radarr:    radarr.New(gctx.Crate().Sqlite.Query()),
		Sonarr:    sonarr.New(gctx.Crate().Sqlite.Query()),
		Jellystat: jellystat.New(gctx),
	}
}
