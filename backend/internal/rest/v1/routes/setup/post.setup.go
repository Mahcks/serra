package setup

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

type SetupRequest struct {
	Type             string             `json:"type" validate:"required,oneof=emby jellyfin"`
	URL              string             `json:"url" validate:"required,url"`
	APIKey           string             `json:"api_key"`
	RequestSystem    string             `json:"request_system" validate:"required,oneof=built_in external"`
	RequestSystemURL string             `json:"request_system_url,omitempty" validate:"omitempty,url"`
	Radarr           []ArrServiceConfig `json:"radarr,omitempty"`
	Sonarr           []ArrServiceConfig `json:"sonarr,omitempty"`
}

type ArrServiceConfig struct {
	Name                string `json:"name"`
	BaseURL             string `json:"base_url" validate:"required,url"`
	APIKey              string `json:"api_key" validate:"required"`
	QualityProfile      string `json:"quality_profile" validate:"required"`
	RootFolderPath      string `json:"root_folder_path" validate:"required"`
	MinimumAvailability string `json:"minimum_availability" validate:"required,oneof=announced released in_cinemas"`
}

func (rg *RouteGroup) Initialize(ctx *respond.Ctx) error {
	var req SetupRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Failed to parse request body")
	}

	// Validate request system URL is provided when using external system
	if req.RequestSystem == string(structures.RequestSystemExternal) && req.RequestSystemURL == "" {
		return apiErrors.ErrBadRequest().SetDetail("Request system URL is required when using external request system")
	}

	// Check if setup is already complete
	_, err := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingSetupComplete.String())
	if err == nil {
		return apiErrors.ErrBadRequest().SetDetail("Setup has already been completed")
	}

	// Begin transaction
	tx, err := rg.gctx.Crate().Sqlite.DB().Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get queries with transaction
	txQueries := repository.New(tx)

	// Store media server settings
	if req.RequestSystemURL != "" {
		req.RequestSystemURL = ""
	}

	settings := map[structures.Setting]string{
		structures.SettingMediaServerType:   req.Type,
		structures.SettingMediaServerURL:    req.URL,
		structures.SettingMediaServerAPIKey: req.APIKey,
		structures.SettingRequestSystem:     req.RequestSystem,
		structures.SettingRequestSystemURL:  req.RequestSystemURL,
		structures.SettingSetupComplete:     "true",
	}

	// Add request system URL if using external system
	if req.RequestSystem == string(structures.RequestSystemExternal) {
		settings[structures.SettingRequestSystemURL] = req.RequestSystemURL
	}

	// Insert or update settings
	for key, value := range settings {
		if err := txQueries.UpsertSetting(ctx.Context(), repository.UpsertSettingParams{
			Key:   key.String(),
			Value: value,
		}); err != nil {
			return fmt.Errorf("failed to store setting %s: %w", key, err)
		}
	}

	// Loop over and create arr_services
	for _, radarr := range req.Radarr {
		err := txQueries.CreateArrService(ctx.Context(), repository.CreateArrServiceParams{
			ID:                  uuid.NewString(),
			Type:                "radarr",
			Name:                radarr.Name,
			BaseUrl:             radarr.BaseURL,
			ApiKey:              radarr.APIKey,
			QualityProfile:      radarr.QualityProfile,
			RootFolderPath:      radarr.RootFolderPath,
			MinimumAvailability: radarr.MinimumAvailability,
		})
		if err != nil {
			return apiErrors.ErrInternalServerError().SetDetail(fmt.Sprintf("Failed to create Radarr service: %v", err))
		}
	}

	for _, sonarr := range req.Sonarr {
		err := txQueries.CreateArrService(ctx.Context(), repository.CreateArrServiceParams{
			ID:                  uuid.NewString(),
			Type:                "sonarr",
			Name:                sonarr.Name,
			BaseUrl:             sonarr.BaseURL,
			ApiKey:              sonarr.APIKey,
			QualityProfile:      sonarr.QualityProfile,
			RootFolderPath:      sonarr.RootFolderPath,
			MinimumAvailability: sonarr.MinimumAvailability,
		})
		if err != nil {
			return apiErrors.ErrInternalServerError().SetDetail(fmt.Sprintf("Failed to create Sonarr service: %v", err))
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit settings: %w", err)
	}

	// Reload the config service to reflect the new settings
	if err := rg.gctx.Crate().Config.Reload(ctx.Context()); err != nil {
		return fmt.Errorf("failed to reload configuration: %w", err)
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "Initial setup completed successfully",
	})
}
