package services

import (
	"github.com/mahcks/serra/internal/services/auth"
	"github.com/mahcks/serra/internal/services/configservice"
	"github.com/mahcks/serra/internal/services/radarr"
	"github.com/mahcks/serra/internal/services/sonarr"
	"github.com/mahcks/serra/internal/services/sqlite"
)

type Crate struct {
	Config      *configservice.Service
	Sqlite      sqlite.Service
	AuthService auth.Authmen
	Radarr      radarr.Service
	Sonarr      sonarr.Service
}
