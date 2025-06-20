package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Bootstrap is the minimal config needed before DB is ready
type Bootstrap struct {
	Version string `mapstructure:"version" json:"version"`

	REST struct {
		Address string `mapstructure:"address" json:"address"`
		Port    string `mapstructure:"port" json:"port"`
	} `mapstructure:"rest" json:"rest"`

	SQLite struct {
		Path string `mapstructure:"path" json:"path"`
	} `mapstructure:"sqlite" json:"sqlite"`

	Credentials struct {
		JwtSecret string `mapstructure:"jwt_secret" json:"jwt_secret"`
	} `mapstructure:"credentials" json:"credentials"`
}

// NewBootstrap loads only what's needed to initialize DB
func NewBootstrap(version string) (*Bootstrap, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.AddConfigPath("./config")
	v.AddConfigPath("/home/nonroot/config")

	if version == "dev" {
		v.SetConfigName("config.dev.yaml")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("No bootstrap config file found, using only ENV variables")
		} else {
			return nil, fmt.Errorf("bootstrap config load error: %w", err)
		}
	}

	// Bind only what's necessary to start
	v.BindEnv("http.address")
	v.BindEnv("http.ports.rest")
	v.BindEnv("sqlite.path")
	v.BindEnv("credentials.jwt_secret")

	c := &Bootstrap{}
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}
	c.Version = version

	return c, nil
}
