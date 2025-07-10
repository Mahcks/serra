package config

import (
	"fmt"
	"strings"

	"github.com/mahcks/serra/pkg/structures"
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	SetupComplete structures.Setting `mapstructure:"setup_complete"`

	MediaServer struct {
		Type   structures.Provider `mapstructure:"type"`
		URL    structures.Setting  `mapstructure:"url"`
		APIKey structures.Setting  `mapstructure:"api_key"`
	} `mapstructure:"media_server"`

	Jellystat struct {
		URL    structures.Setting `mapstructure:"url"`
		APIKey structures.Setting `mapstructure:"api_key"`
	} `mapstructure:"jellystat"`

	TMDB struct {
		APIKey structures.Setting `mapstructure:"api_key"`
	} `mapstructure:"tmdb"`
}

// New creates a new Config instance with the given settings
func New(settings map[string]interface{}) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	v.SetDefault("setup_complete", "")

	v.SetDefault("media_server.type", "")
	v.SetDefault("media_server.url", "")
	v.SetDefault("media_server.api_key", "")

	v.SetDefault("jellystat.url", "")
	v.SetDefault("jellystat.api_key", "")

	v.SetDefault("tmdb.api_key", "")

	// Convert flat keys to nested structure
	nestedSettings := make(map[string]interface{})
	for key, value := range settings {

		switch key {
		case structures.SettingSetupComplete.String():
			nestedSettings["setup_complete"] = value
		case structures.SettingMediaServerType.String():
			nestedSettings["media_server.type"] = value
		case structures.SettingMediaServerURL.String():
			nestedSettings["media_server.url"] = value
		case structures.SettingMediaServerAPIKey.String():
			nestedSettings["media_server.api_key"] = value
		case structures.SettingJellystatURL.String():
			nestedSettings["jellystat.url"] = value
		case structures.SettingJellystatAPIKey.String():
			nestedSettings["jellystat.api_key"] = value
		case structures.SettingTMDBAPIKey.String():
			nestedSettings["tmdb.api_key"] = value
		}
	}

	// Set the values in Viper
	for key, value := range nestedSettings {
		v.Set(key, value)
	}

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// NewEmpty creates a new empty Config instance with default values
func NewEmpty() *Config {
	cfg, _ := New(make(map[string]interface{}))
	return cfg
}

// Get returns a value from the config using a dot-notation path
func (c *Config) Get(path string) interface{} {
	parts := strings.Split(path, ".")
	current := interface{}(c)

	for _, part := range parts {
		fmt.Println("Current part:", part)
		switch v := current.(type) {
		case *Config:
			switch part {
			case "setup_complete":
				return v.SetupComplete
			case "media_server":
				current = v.MediaServer
			}
		case struct {
			Type   structures.Setting
			URL    structures.Setting
			APIKey structures.Setting
		}:
			switch part {
			case "type":
				return v.Type
			case "url":
				return v.URL
			case "api_key":
				return v.APIKey
			}
		}
	}

	return nil
}
