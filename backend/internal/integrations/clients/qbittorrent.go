package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mahcks/serra/pkg/downloadclient"
	"github.com/mahcks/serra/utils"
)

// QBitTorrentClient implements the DownloadClientInterface for qBittorrent
type QBitTorrentClient struct {
	config     downloadclient.Config
	httpClient *http.Client
	sid        string // Session ID for authentication
	connected  bool
	lastError  string
}

// NewQBitTorrentClient creates a new qBittorrent client
func NewQBitTorrentClient(config downloadclient.Config) (downloadclient.Interface, error) {
	client := &QBitTorrentClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	return client, nil
}

// GetType returns the client type
func (c *QBitTorrentClient) GetType() string {
	return "qbittorrent"
}

// GetName returns the client name
func (c *QBitTorrentClient) GetName() string {
	return c.config.Name
}

// Connect establishes a connection to qBittorrent
func (c *QBitTorrentClient) Connect(ctx context.Context) error {
	if c.config.Username == nil || c.config.Password == nil {
		return fmt.Errorf("username and password are required for qBittorrent")
	}

	// Build the base URL
	scheme := utils.Ternary(c.config.UseSSL, "https", "http")
	baseURL := fmt.Sprintf("%s://%s:%d", scheme, c.config.Host, c.config.Port)

	// Login to get session ID
	loginURL := utils.BuildURL(baseURL, "/api/v2/auth/login", nil)
	form := url.Values{}
	form.Set("username", utils.DerefString(c.config.Username))
	form.Set("password", utils.DerefString(c.config.Password))

	req, err := http.NewRequestWithContext(ctx, "POST", loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		c.lastError = err.Error()
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", baseURL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.lastError = err.Error()
		return err
	}
	defer resp.Body.Close()

	if !utils.IsHTTPSuccess(resp.StatusCode) {
		body, _ := io.ReadAll(resp.Body)
		c.lastError = fmt.Sprintf("login failed: %s", string(body))
		return fmt.Errorf("qBittorrent login failed: %s", resp.Status)
	}

	// Extract session ID from cookies
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "SID" {
			c.sid = cookie.Value
			break
		}
	}

	if c.sid == "" {
		c.lastError = "no session ID received"
		return fmt.Errorf("failed to get session ID from qBittorrent")
	}

	c.connected = true
	c.lastError = ""
	slog.Debug("Connected to qBittorrent", "host", c.config.Host, "port", c.config.Port)
	return nil
}

// Disconnect closes the connection
func (c *QBitTorrentClient) Disconnect(ctx context.Context) error {
	if c.connected {
		// Logout from qBittorrent
		scheme := utils.Ternary(c.config.UseSSL, "https", "http")
		baseURL := fmt.Sprintf("%s://%s:%d", scheme, c.config.Host, c.config.Port)
		logoutURL := utils.BuildURL(baseURL, "/api/v2/auth/logout", nil)

		req, err := http.NewRequestWithContext(ctx, "POST", logoutURL, nil)
		if err == nil {
			if c.sid != "" {
				req.AddCookie(&http.Cookie{Name: "SID", Value: c.sid})
			}
			req.Header.Set("Referer", baseURL)
			c.httpClient.Do(req) // Ignore errors on logout
		}
	}

	c.connected = false
	c.sid = ""
	return nil
}

// GetDownloads retrieves all active downloads
func (c *QBitTorrentClient) GetDownloads(ctx context.Context) ([]downloadclient.Item, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected to qBittorrent")
	}

	scheme := utils.Ternary(c.config.UseSSL, "https", "http")
	baseURL := fmt.Sprintf("%s://%s:%d", scheme, c.config.Host, c.config.Port)

	// Get torrents info
	torrentsURL := utils.BuildURL(baseURL, "/api/v2/torrents/info", map[string]string{"filter": "downloading"})
	req, err := http.NewRequestWithContext(ctx, "GET", torrentsURL, nil)
	if err != nil {
		c.lastError = err.Error()
		return nil, err
	}

	if c.sid != "" {
		req.AddCookie(&http.Cookie{Name: "SID", Value: c.sid})
	}
	req.Header.Set("Referer", baseURL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.lastError = err.Error()
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.lastError = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return nil, fmt.Errorf("failed to get torrents: %s", resp.Status)
	}

	var torrents []qbitTorrentInfo
	if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
		c.lastError = err.Error()
		return nil, err
	}

	// Convert to DownloadItem format
	var downloads []downloadclient.Item
	for _, torrent := range torrents {
		download := downloadclient.Item{
			ID:       torrent.Hash,
			Name:     torrent.Name,
			Hash:     torrent.Hash,
			Progress: torrent.Progress * 100, // Convert from 0-1 to 0-100
			Status:   torrent.State,
			TimeLeft: formatTimeLeft(torrent.ETA),
			ETA:      int64(torrent.ETA),
			AddedOn:  time.Unix(torrent.AddedOn, 0),
		}
		downloads = append(downloads, download)
	}

	return downloads, nil
}

// GetDownloadProgress retrieves progress for a specific download
func (c *QBitTorrentClient) GetDownloadProgress(ctx context.Context, downloadID string) (*downloadclient.Progress, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected to qBittorrent")
	}

	// Get specific torrent info
	scheme := "http"
	if c.config.UseSSL {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s:%d", scheme, c.config.Host, c.config.Port)

	torrentsURL := fmt.Sprintf("%s/api/v2/torrents/info?hashes=%s", baseURL, downloadID)
	req, err := http.NewRequestWithContext(ctx, "GET", torrentsURL, nil)
	if err != nil {
		c.lastError = err.Error()
		return nil, err
	}

	if c.sid != "" {
		req.AddCookie(&http.Cookie{Name: "SID", Value: c.sid})
	}
	req.Header.Set("Referer", baseURL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.lastError = err.Error()
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.lastError = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return nil, fmt.Errorf("failed to get torrent info: %s", resp.Status)
	}

	var torrents []qbitTorrentInfo
	if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
		c.lastError = err.Error()
		return nil, err
	}

	if len(torrents) == 0 {
		return nil, fmt.Errorf("torrent not found: %s", downloadID)
	}

	torrent := torrents[0]
	progress := &downloadclient.Progress{
		Progress: torrent.Progress * 100,
		TimeLeft: formatTimeLeft(torrent.ETA),
		Status:   torrent.State,
	}

	return progress, nil
}

// IsConnected returns whether the client is connected
func (c *QBitTorrentClient) IsConnected() bool {
	return c.connected
}

// GetConnectionInfo returns connection details
func (c *QBitTorrentClient) GetConnectionInfo() downloadclient.ConnectionInfo {
	return downloadclient.ConnectionInfo{
		Host:      c.config.Host,
		Port:      c.config.Port,
		UseSSL:    c.config.UseSSL,
		Connected: c.connected,
		LastError: c.lastError,
	}
}

// qbitTorrentInfo represents the qBittorrent API response structure
type qbitTorrentInfo struct {
	Hash     string  `json:"hash"`
	Name     string  `json:"name"`
	Progress float64 `json:"progress"`
	State    string  `json:"state"`
	ETA      int     `json:"eta"`
	AddedOn  int64   `json:"added_on"`
	Category string  `json:"category"`
	Tags     string  `json:"tags"`
}

// formatTimeLeft formats ETA in seconds to human-readable format
func formatTimeLeft(eta int) string {
	if eta <= 0 {
		return ""
	}

	hours := eta / 3600
	minutes := (eta % 3600) / 60
	seconds := eta % 60

	var result string
	if hours > 0 {
		result += fmt.Sprintf("%dh", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%dm", minutes)
	}
	if seconds > 0 || result == "" {
		result += fmt.Sprintf("%ds", seconds)
	}

	return result
}

