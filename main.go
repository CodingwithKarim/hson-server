package main

import (
	"context"
	"fmt"
	"hson-server/internal/app"
	"hson-server/internal/config"
	"hson-server/internal/logger"
	"hson-server/internal/router"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Parse command-line flags to get the HSON file path, server port to listen on, live-reloading option, and auth settings
	cfg, err := config.Load()

	if err != nil {
		fmt.Printf("failed to load configuration from environment or flags: %v\n", err)
		os.Exit(1)
	}

	// Setup logger singleton that can be accessed by entire app
	logger.SetupSingleton(cfg.LogLevel, cfg.IsLogVerbose)

	// Init the app struct
	app := &app.App{
		Data:     map[string]any{},
		FilePath: cfg.DBPath,
	}

	// Load data from the HSON file into memory / app.Data
	if err := app.LoadDataFromFile(); err != nil {
		logger.Fatal("Failed to access the database file", "path", cfg.DBPath, "err", err)
		os.Exit(1)
	}

	// Only watch HSON / data file for updates if live reload was requested
	if cfg.LiveReload {
		go config.WatchHSONFile(app)
		logger.Info("Live‐reload enabled: watching", "file", cfg.DBPath)
	}

	// Init HTTP router / handler that handles incoming requests and dispatches actions based on HTTP verb
	handler := router.NewHTTPHandler(app, cfg.Auth)

	// Init a HTTP server using the specified port and router
	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: handler,
	}

	// Start the HTTP server in a background goroutine so can handle shutdown signals below
	go func() {
		logger.Info("Starting HSON Server", "port", cfg.ServerPort, "data file", cfg.DBPath)

		// Start serving HTTP requests
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HSON Server failed to listen and serve", "port", cfg.ServerPort, "err", err)
			os.Exit(1)
		}
	}()

	// Create a channel to receive signals
	stop := make(chan os.Signal, 1)

	// Hookup up channel for interupt signal (CTRL + C or killing process)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Block until channel receives an interrupt or kill signal
	<-stop

	logger.Info("Shutdown signal received, shutting down...")

	// Attempt graceful shutdown
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Error("Graceful shutdown failed, forcing exit", "err", err)
	} else {
		logger.Info("🌙  HSON Server shutdown complete. See you next time!")
	}
}
