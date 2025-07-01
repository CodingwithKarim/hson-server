package main

import (
	"context"
	"flag"
	"hson-server/internal/app"
	"hson-server/internal/logger"
	"hson-server/internal/router"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	"github.com/fsnotify/fsnotify"
)

func main() {
	// Parse command-line flags to get the HSON file path, server port to listen on, and live-reloading option
	dbPath, serverPort, liveReload := parseAppFlags()

	// Setup logger singleton that can be accessed by entire app
	logger.Setup()

	// Init the app struct
	app := &app.App{
		Data:     map[string]any{},
		FilePath: dbPath,
	}

	// Load data from the HSON file into memory / app.Data
	if err := app.LoadDataFromFile(); err != nil {
		logger.Fatal("Failed to access the database file", "path", dbPath, "err", err)
	}

	// Only watch HSON / data file for updates if live reload was requested
	if liveReload {
		go watchHSONFile(app)
		logger.Info("Live‐reload enabled: watching", "file", dbPath)
	}

	// Init HTTP router / handler that handles incoming requests and dispatches actions based on HTTP verb
	handler := router.NewHTTPHandler(app)

	// Init a HTTP server using the specified port and router
	server := &http.Server{
		Addr:    ":" + serverPort,
		Handler: handler,
	}

	// Start the HTTP server in a background goroutine so can handle shutdown signals below
	go func() {
		logger.Info("Starting HSON Server", "port", serverPort, "data file", dbPath)

		// Start serving HTTP requests
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HSON Server failed to listen and serve", "port", serverPort, "err", err)
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

func parseAppFlags() (dbPath, serverPort string, liveReload bool) {
	// Register cli flags for configuring server e.g: port, hson file path, live-reloading, etc...
	flag.StringVar(&dbPath, "db", "data.hson", "path to your HSON database file")
	flag.StringVar(&dbPath, "database", "data.hson", "alias for --db")
	flag.StringVar(&serverPort, "port", "3000", "port the server will listen on")
	flag.BoolVar(&liveReload, "live-reload", false, "watch HSON file and reload on external changes")

	// Register cli flags for logger e.g: log level, verbose option
	logger.RegisterFlags()

	// Parse all registered command-line flags
	flag.Parse()

	return
}

func watchHSONFile(app *app.App) {
	// Init the live reload watcher
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		logger.Error("Live Reload Watcher initialization failed", "err", err)
		return
	}

	defer watcher.Close()

	// Start monitoring file path using watcher
	if err := watcher.Add(app.FilePath); err != nil {
		logger.Error("Watcher.Add failed", "path", app.FilePath, "err", err)
		return
	}

	// Loop through the watcher events indefintely
	for ev := range watcher.Events {
		// Only monitor write events and ensure update did not come from code / app.Persist() call
		if ev.Op&fsnotify.Write == 0 || atomic.LoadUint32(&app.SelfWriting) == 1 {
			continue
		}

		logger.Info("Reloading HSON from disk")

		// Load data from HSON file to app memory
		if err := app.LoadDataFromFile(); err != nil {
			logger.Error("Reload failed", "err", err)
		}
	}

	if err := <-watcher.Errors; err != nil {
		logger.Error("Watcher error", "err", err)
	}
}
