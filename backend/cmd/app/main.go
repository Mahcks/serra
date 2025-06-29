package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mahcks/serra/config"
	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/internal/integrations"
	"github.com/mahcks/serra/internal/rest"
	"github.com/mahcks/serra/internal/services/auth"
	"github.com/mahcks/serra/internal/services/configservice"
	"github.com/mahcks/serra/internal/services/sqlite"
)

var (
	Version   = "dev"
	Timestamp = "unknown"
)

func main() {
	Timestamp := time.Now().Format(time.RFC3339)

	if v := os.Getenv("VERSION"); v != "" {
		Version = v
	}

	// Set the log level based on the version
	var logLevel slog.Level
	if Version == "dev" {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	if Version == "dev" {
		slog.Info("Starting in development mode", "version", Version, "timestamp", Timestamp)
	} else {
		slog.Info("Starting in production mode", "version", Version, "timestamp", Timestamp)
	}

	bootstrap, err := config.NewBootstrap(Version)
	if err != nil {
		slog.Error("Failed to load bootstrap configuration", "error", err)
		os.Exit(1)
	}

	// Create global context with all services
	gctx, cancel := global.WithCancel(global.New(
		context.Background(),
		bootstrap,
		Version,
		Timestamp,
	))

	{
		// Initialize SQLite first
		slog.Info("sqlite", "status", "starting")
		gctx.Crate().Sqlite, err = sqlite.Setup(gctx, Version, sqlite.SetupOptions{
			Path: gctx.Bootstrap().SQLite.Path,
		})
		if err != nil {
			slog.Error("Failed to initialize SQLite service", "error", err)
			os.Exit(1)
		}
		slog.Info("setup service", "service", "sqlite")
	}

	{
		// Initialize config service
		slog.Info("config", "status", "starting")
		gctx.Crate().Config = configservice.New(gctx.Crate().Sqlite.Query())
		if err := gctx.Crate().Config.Load(context.Background()); err != nil {
			slog.Error("Failed to load configuration", "error", err)
			os.Exit(1)
		}
		slog.Info("setup service", "service", "config")
	}

	{
		// Initialize authentication service
		gctx.Crate().AuthService = auth.New(
			gctx.Bootstrap().Credentials.JwtSecret,
			"localhost",
			true,
		)
		slog.Info("setup service", "service", "auth")
	}

	// Initialize integration services
	ints := integrations.New(gctx)
	downloadPoller, _ := integrations.NewDownloadPoller(gctx, integrations.DownloadPollerOptions{})
	slog.Info("setup service", "service", "integrations")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	done := make(chan struct{})
	wg := sync.WaitGroup{}

	go func() {
		<-interrupt
		cancel()

		go func() {
			select {
			case <-time.After(time.Minute):
			case <-interrupt:
			}
			slog.Warn("Force shutdown after timeout")
		}()

		slog.Warn("Shutting down...")

		wg.Wait()

		if gctx.Crate() != nil && gctx.Crate().Sqlite != nil {
			if err := gctx.Crate().Sqlite.Close(); err != nil {
				slog.Error("Error closing sqlite connection", "error", err)
			}
		}

		downloadPoller.Stop(gctx)

		close(done)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		slog.Info("rest api", "status", "starting")
		if err := rest.New(gctx, ints); err != nil {
			slog.Error("Failed to start rest api", "error", err)
			os.Exit(1)
		}
		slog.Info("rest api", "status", "initialized")
	}()

	<-done
	slog.Info("Shutdown complete")
	os.Exit(0)
}
