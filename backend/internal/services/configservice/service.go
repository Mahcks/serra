package configservice

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mahcks/serra/config"
	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/pkg/structures"
)

type Service struct {
	cfg     *config.Config
	queries *repository.Queries
}

func New(queries *repository.Queries) *Service {
	return &Service{
		cfg:     config.NewEmpty(),
		queries: queries,
	}
}

// Load loads all settings from the database into the config
func (s *Service) Load(ctx context.Context) error {
	settings, err := s.queries.GetAllSettings(ctx)
	if err != nil {
		return fmt.Errorf("failed to get settings: %w", err)
	}

	// Create a map of settings
	settingsMap := make(map[string]interface{})
	for _, setting := range settings {
		var value interface{}
		if err := json.Unmarshal([]byte(setting.Value), &value); err != nil {
			// If not JSON, use as string
			value = setting.Value
		}

		// Convert value to string
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case bool:
			strValue = strconv.FormatBool(v)
		case int:
			strValue = strconv.Itoa(v)
		case float64:
			strValue = strconv.FormatFloat(v, 'f', -1, 64)
		default:
			// For any other type, convert to JSON string
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("failed to marshal value: %w", err)
			}
			strValue = string(jsonBytes)
		}

		settingsMap[setting.Key] = strValue
	}

	// Create new config from settings
	newCfg, err := config.New(settingsMap)
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	s.cfg = newCfg
	return nil
}

// Get returns the current configuration
func (s *Service) Get() *config.Config {
	return s.cfg
}

// Reload reloads the configuration from the database
func (s *Service) Reload(ctx context.Context) error {
	return s.Load(ctx)
}

// Set sets a setting value using the Setting type
func (s *Service) Set(ctx context.Context, key structures.Setting, value interface{}) error {
	// Convert value to string
	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	case bool:
		strValue = strconv.FormatBool(v)
	case int:
		strValue = strconv.Itoa(v)
	case float64:
		strValue = strconv.FormatFloat(v, 'f', -1, 64)
	default:
		// For any other type, convert to JSON string
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		strValue = string(jsonBytes)
	}

	// Store in database
	err := s.queries.UpsertSetting(ctx, repository.UpsertSettingParams{
		Key:   key.String(),
		Value: strValue,
	})
	if err != nil {
		return fmt.Errorf("failed to store setting: %w", err)
	}

	// Reload all settings from database to ensure consistency
	if err := s.Load(ctx); err != nil {
		return fmt.Errorf("failed to reload configuration: %w", err)
	}

	return nil
}
