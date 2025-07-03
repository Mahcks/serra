package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mahcks/serra/pkg/downloadclient"
	"github.com/mahcks/serra/utils"
)

// SABnzbdClient implements the DownloadClientInterface for SABnzbd
type SABnzbdClient struct {
	config     downloadclient.Config
	httpClient *http.Client
	connected  bool
	lastError  string
}

// NewSABnzbdClient creates a new SABnzbd client
func NewSABnzbdClient(config downloadclient.Config) (downloadclient.Interface, error) {
	client := &SABnzbdClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	return client, nil
}

// GetType returns the client type
func (c *SABnzbdClient) GetType() string {
	return "sabnzbd"
}

// GetName returns the client name
func (c *SABnzbdClient) GetName() string {
	return c.config.Name
}

// Connect establishes a connection to SABnzbd
func (c *SABnzbdClient) Connect(ctx context.Context) error {
	if c.config.APIKey == nil {
		return fmt.Errorf("API key is required for SABnzbd")
	}

	// Test connection by getting queue info
	_, err := c.getQueueInfo(ctx)
	if err != nil {
		c.lastError = err.Error()
		return err
	}

	c.connected = true
	c.lastError = ""
	slog.Debug("Connected to SABnzbd", "host", c.config.Host, "port", c.config.Port)
	return nil
}

// Disconnect closes the connection
func (c *SABnzbdClient) Disconnect(ctx context.Context) error {
	c.connected = false
	return nil
}

// GetDownloads retrieves all active downloads
func (c *SABnzbdClient) GetDownloads(ctx context.Context) ([]downloadclient.Item, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected to SABnzbd")
	}

	queueInfo, err := c.getQueueInfo(ctx)
	if err != nil {
		c.lastError = err.Error()
		return nil, err
	}

	var downloads []downloadclient.Item
	for _, slot := range queueInfo.Queue.Slots {
		// Include all downloads, not just active ones
		// The modular poller will handle filtering if needed

		progress := utils.SafeAtof(slot.Percentage)
		size := utils.SafeAtoi64(slot.Size)
		sizeLeft := utils.SafeAtoi64(slot.SizeLeft)

		// Map SABnzbd status to a more generic status
		status := c.mapSABnzbdStatus(slot.Status)

		// Generate a meaningful name based on available data

		download := downloadclient.Item{
			ID:            slot.NzbID,
			Name:          slot.FileName,
			Progress:      progress,
			Status:        status,
			Size:          size,
			SizeLeft:      sizeLeft,
			TimeLeft:      formatSABTimeLeft(slot.TimeLeft),
			DownloadSpeed: formatSABSpeed(slot.MB),
			Category:      slot.Category,
			AddedOn:       time.Unix(slot.AddedOn, 0),
		}
		downloads = append(downloads, download)
	}

	return downloads, nil
}

// GetDownloadProgress retrieves progress for a specific download
func (c *SABnzbdClient) GetDownloadProgress(ctx context.Context, downloadID string) (*downloadclient.Progress, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected to SABnzbd")
	}

	queueInfo, err := c.getQueueInfo(ctx)
	if err != nil {
		c.lastError = err.Error()
		return nil, err
	}

	for _, slot := range queueInfo.Queue.Slots {
		if slot.NzbID == downloadID {
			progress := utils.SafeAtof(slot.Percentage)
			return &downloadclient.Progress{
				Progress:      progress,
				TimeLeft:      formatSABTimeLeft(slot.TimeLeft),
				DownloadSpeed: formatSABSpeed(slot.MB),
				Status:        slot.Status,
			}, nil
		}
	}

	return nil, fmt.Errorf("download not found: %s", downloadID)
}

// IsConnected returns whether the client is connected
func (c *SABnzbdClient) IsConnected() bool {
	return c.connected
}

// GetConnectionInfo returns connection details
func (c *SABnzbdClient) GetConnectionInfo() downloadclient.ConnectionInfo {
	return downloadclient.ConnectionInfo{
		Host:      c.config.Host,
		Port:      c.config.Port,
		UseSSL:    c.config.UseSSL,
		Connected: c.connected,
		LastError: c.lastError,
	}
}

// getQueueInfo fetches queue information from SABnzbd
func (c *SABnzbdClient) getQueueInfo(ctx context.Context) (*sabnzbdQueueResponse, error) {
	scheme := utils.Ternary(c.config.UseSSL, "https", "http")
	baseURL := fmt.Sprintf("%s://%s:%d", scheme, c.config.Host, c.config.Port)
	url := utils.BuildURL(baseURL, "/api", map[string]string{
		"mode":   "queue",
		"output": "json",
		"apikey": utils.DerefString(c.config.APIKey),
	})

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SABnzbd API request failed: %s", resp.Status)
	}

	var result sabnzbdQueueResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// sabnzbdQueueResponse represents the SABnzbd queue API response
type sabnzbdQueueResponse struct {
	Queue struct {
		Slots []sabnzbdSlot `json:"slots"`
	} `json:"queue"`
}

// sabnzbdSlot represents a download slot in SABnzbd
type sabnzbdSlot struct {
	NzbID      string `json:"nzo_id"`
	FileName   string `json:"filename"`
	Percentage string `json:"percentage"`
	TimeLeft   string `json:"timeleft"`
	Size       string `json:"size"`
	SizeLeft   string `json:"sizeleft"`
	Status     string `json:"status"`
	MB         string `json:"mb"` // Download speed in MB/s
	Category   string `json:"cat"`
	AddedOn    int64  `json:"added_on"`
}

// formatSABTimeLeft formats SABnzbd time format to human-readable format
func formatSABTimeLeft(raw string) string {
	if raw == "" || raw == "0:00:00" {
		return ""
	}

	parts := strings.Split(raw, ":")
	if len(parts) != 3 {
		return raw
	}

	hours := parts[0]
	minutes := parts[1]
	seconds := parts[2]

	var result string
	if hours != "0" {
		result += fmt.Sprintf("%sh", hours)
	}
	if minutes != "0" {
		result += fmt.Sprintf("%sm", minutes)
	}
	if seconds != "0" || result == "" {
		result += fmt.Sprintf("%ss", seconds)
	}

	return result
}

// formatSABSpeed formats SABnzbd speed format
func formatSABSpeed(mbStr string) string {
	if mbStr == "" || mbStr == "0" {
		return ""
	}

	mb, err := strconv.ParseFloat(mbStr, 64)
	if err != nil {
		return ""
	}

	if mb >= 1024 {
		return fmt.Sprintf("%.1f GB/s", mb/1024)
	}
	return fmt.Sprintf("%.1f MB/s", mb)
}

// mapSABnzbdStatus maps SABnzbd-specific statuses to generic ones
func (c *SABnzbdClient) mapSABnzbdStatus(sabStatus string) string {
	switch strings.ToLower(sabStatus) {
	case "downloading":
		return "downloading"
	case "queued":
		return "queued"
	case "paused":
		return "paused"
	case "completed":
		return "completed"
	case "extracting":
		return "extracting"
	case "verifying":
		return "verifying"
	case "repairing":
		return "repairing"
	case "failed":
		return "failed"
	default:
		return sabStatus
	}
}
