package routes

import (
	"fmt"
	"strconv"
	"time"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/pkg/structures"
)

var uptime = time.Now()

type HealthResponse struct {
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
}

type SetupStatusResponse struct {
	SetupComplete bool `json:"setup_complete"`
}

func (rg *RouteGroup) Index(ctx *respond.Ctx) error {
	fmt.Println(rg.Config())

	return ctx.JSON(HealthResponse{
		Version: rg.gctx.Bootstrap().Version,
		Uptime:  strconv.Itoa(int(uptime.UnixMilli())),
	})
}

// SetupStatus checks if the initial setup has been completed
func (rg *RouteGroup) SetupStatus(ctx *respond.Ctx) error {
	// Check if setup is complete by looking for the setting
	_, err := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingSetupComplete.String())
	setupComplete := err == nil // If we can find the setting, setup is complete

	return ctx.JSON(SetupStatusResponse{
		SetupComplete: setupComplete,
	})
}
