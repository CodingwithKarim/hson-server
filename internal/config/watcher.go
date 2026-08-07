package config

import (
	"hson-server/internal/app"
	"hson-server/internal/logger"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
)

func WatchHSONFile(app *app.App) {
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

	for {
		select {
		case err, ok := <-watcher.Errors:
			if !ok {
				logger.Warn("Watcher channel closed, live reload disabled")
				return
			}

			if err != nil {
				logger.Error("Watcher error", "err", err)
			}

		case ev, ok := <-watcher.Events:
			if !ok {
				logger.Warn("Watcher channel closed, live reload disabled")
				return
			}

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
	}
}
