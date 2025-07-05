package routes

import (
	"fmt"
	"strconv"
	"time"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/internal/websocket"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
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
	testArr, err := rg.integrations.Jellystat.GetLibraryOverview()
	if err != nil {
		return apiErrors.ErrInternalServerError()
	}

	fmt.Println(testArr)

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

// TestWebSocket sends a test WebSocket message for debugging
func (rg *RouteGroup) TestWebSocket(ctx *respond.Ctx) error {
	// Create test download data
	testDownloads := []structures.DownloadProgressPayload{
		{
			ID:           "test-download-1",
			Title:        "Test Movie 2024",
			TorrentTitle: "Test.Movie.2024.1080p.BluRay.x264-GROUP",
			Source:       "radarr",
			Hash:         "test-hash-123",
			Progress:     75.5,
			TimeLeft:     "15m 30s",
			Status:       "downloading",
			LastUpdated:  time.Now().Format(time.RFC3339),
		},
		{
			ID:           "test-download-2", 
			Title:        "Test Series S01E01",
			TorrentTitle: "Test.Series.S01E01.1080p.WEB-DL.x264-GROUP",
			Source:       "sonarr",
			Hash:         "test-hash-456",
			Progress:     25.0,
			TimeLeft:     "1h 15m",
			Status:       "downloading", 
			LastUpdated:  time.Now().Format(time.RFC3339),
		},
	}
	
	// Get connected client count for response
	connectedClients := websocket.GetConnectionCount()
	connectedUsers := websocket.GetConnectedUsers()
	
	// Broadcast the test data
	websocket.BroadcastDownloadProgressBatch(testDownloads)
	
	return ctx.JSON(map[string]interface{}{
		"status": "test WebSocket message sent",
		"downloads": len(testDownloads),
		"connectedClients": connectedClients,
		"connectedUsers": connectedUsers,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
