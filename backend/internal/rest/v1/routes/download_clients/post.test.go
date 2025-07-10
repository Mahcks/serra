package downloadclients

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/serra/internal/integrations/clients"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/downloadclient"
)

type TestRequest struct {
	Type     string `json:"type" validate:"required,oneof=qbittorrent sabnzbd"`
	Host     string `json:"host" validate:"required"`
	Port     int    `json:"port" validate:"required,min=1,max=65535"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	APIKey   string `json:"api_key,omitempty"`
	UseSSL   bool   `json:"use_ssl"`
}

func (rg *RouteGroup) TestConnection(ctx *respond.Ctx) error {
	var req TestRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Failed to parse request body")
	}

	// Validate required fields based on client type
	if req.Type == "qbittorrent" {
		if req.Username == "" || req.Password == "" {
			return apiErrors.ErrBadRequest().SetDetail("Username and password are required for qBittorrent")
		}
	} else if req.Type == "sabnzbd" {
		if req.APIKey == "" {
			return apiErrors.ErrBadRequest().SetDetail("API key is required for SABnzbd")
		}
	}

	// Create download client config
	config := downloadclient.Config{
		Type:   req.Type,
		Host:   req.Host,
		Port:   req.Port,
		UseSSL: req.UseSSL,
	}

	// Set authentication based on client type
	if req.Username != "" {
		config.Username = &req.Username
	}
	if req.Password != "" {
		config.Password = &req.Password
	}
	if req.APIKey != "" {
		config.APIKey = &req.APIKey
	}

	// Create client instance
	var client downloadclient.Interface
	var err error

	switch req.Type {
	case "qbittorrent":
		client, err = clients.NewQBitTorrentClient(config)
	case "sabnzbd":
		client, err = clients.NewSABnzbdClient(config)
	default:
		return apiErrors.ErrBadRequest().SetDetail("Unsupported client type")
	}

	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to create download client: " + err.Error())
	}

	// Test connection with timeout
	testCtx, cancel := context.WithTimeout(ctx.Context(), 30*time.Second)
	defer cancel()

	if err := client.Connect(testCtx); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Connection failed: " + err.Error())
	}

	// Disconnect after test
	defer client.Disconnect(testCtx)

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "Successfully connected to download client",
		"type":    req.Type,
	})
}