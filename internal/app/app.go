package app

import (
	"hson-server/internal/datatree"
	"hson-server/internal/logger"
	"net/url"
	"os"
	"sync"
	"sync/atomic"

	"github.com/hjson/hjson-go"
)

type App struct {
	Mutex       sync.RWMutex
	Data        map[string]any
	FilePath    string
	SelfWriting uint32
}

func (app *App) LoadDataFromFile() error {
	// Get raw data from the hson / data file
	raw, err := os.ReadFile(app.FilePath)

	if err != nil {
		return err
	}

	// Init a var we will use to store structured manner
	var data map[string]any

	// Convert raw data to structured data
	if err := hjson.Unmarshal(raw, &data); err != nil {
		return err
	}

	// Add a lock to app data
	app.Mutex.Lock()

	// Defer the unlock of lock on function return
	defer app.Mutex.Unlock()

	// Assign new data to app data
	app.Data = data

	return nil
}

func (app *App) Read(path string) (any, error) {
	// Add a lock to app data
	app.Mutex.RLock()

	// Defer the unlock of lock on function return
	defer app.Mutex.RUnlock()

	// Look up the value from app data at the specified path
	return datatree.Lookup(app.Data, path)
}

func (app *App) Write(path string, newVal any) error {
	// Add a lock to app data
	app.Mutex.Lock()

	// Defer the unlock of lock on function return
	defer app.Mutex.Unlock()

	// Set value at the specified path within app data
	if err := datatree.Set(app.Data, path, newVal); err != nil {
		return err
	}

	// Persist updated data back to data file / disk
	if err := app.persist(); err != nil {
		logger.Error("failed to write file", "err", err)
		return err
	}

	// Return nil on successful write of data
	return nil
}

func (app *App) Patch(path string, patchData map[string]any) error {
	// Add a lock to app data
	app.Mutex.Lock()

	// Defer the unlock of lock on function return
	defer app.Mutex.Unlock()

	// Apply the patch to the value at the specified path in the data tree
	if err := datatree.Patch(app.Data, path, patchData); err != nil {
		return err
	}

	// Persist updated data back to data file / disk
	if err := app.persist(); err != nil {
		logger.Error("failed to write file", "err", err)
		return err
	}

	// Return nil on successful patch of data
	return nil
}

func (app *App) Delete(path string, q url.Values) error {
	// Add a lock to app data
	app.Mutex.Lock()

	// Defer the unlock of lock on function return
	defer app.Mutex.Unlock()

	// If filters / query params are provided, fire bulk delete
	if len(q) > 0 {
		filters := datatree.FlattenFilters(q)

		if err := datatree.BulkDelete(app.Data, path, filters); err != nil {
			return err
		}
	} else {
		// Single delete on path when no filter is provided
		if err := datatree.Delete(app.Data, path); err != nil {
			return err
		}
	}

	// Persist change back to hson file
	return app.persist()
}

func (app *App) persist() error {
	// Set value of self writing to 1 / true
	atomic.StoreUint32(&app.SelfWriting, 1)

	// Inevitably, revert the value of self writing back to 0 / false
	defer atomic.StoreUint32(&app.SelfWriting, 0)

	// Convert app data to HJSON encoded data
	hsonBytes, err := hjson.Marshal(app.Data)

	if err != nil {
		return err
	}

	// Write encoded data back to file at app.FilePath
	return os.WriteFile(app.FilePath, hsonBytes, 0o644)
}
