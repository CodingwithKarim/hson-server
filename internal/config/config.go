package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DBPath       string
	ServerPort   string
	LogLevel     string
	LiveReload   bool
	IsLogVerbose bool
}

func Load() (*Config, error) {
	cfg := parseAppFlags()

	resolvedPath, err := resolveDataFile(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	cfg.DBPath = resolvedPath

	return cfg, nil
}

func parseAppFlags() *Config {
	cfg := &Config{}

	// Register cli flags for configuring server e.g: port, hson file path, live-reloading, etc...
	flag.StringVar(&cfg.DBPath, "db", "data.hson", "path to your HSON database file")
	flag.StringVar(&cfg.DBPath, "database", "data.hson", "alias for --db")
	flag.StringVar(&cfg.ServerPort, "port", "3000", "port the server will listen on")
	flag.BoolVar(&cfg.LiveReload, "live-reload", false, "watch HSON file and reload on external changes")

	flag.StringVar(&cfg.LogLevel, "log-level", "info", "log level: debug, info, warn, error")
	flag.BoolVar(&cfg.IsLogVerbose, "verbose", false, "enable verbose logging (adds file and line number and extra fields)")

	// Parse all registered command-line flags
	flag.Parse()

	return cfg
}

func resolveDataFile(dbPath string) (string, error) {
	if dbPath != "data.hson" {
		return dbPath, nil
	}

	if _, err := os.Stat("data.hson"); err == nil {
		abs, err := filepath.Abs("data.hson")
		if err != nil {
			return "", err
		}
		return abs, nil
	}

	exePath, err := os.Executable()

	if err != nil {
		return "", err
	}

	exeDir := filepath.Dir(exePath)
	fallback := filepath.Join(exeDir, "data.hson")

	if _, err := os.Stat(fallback); err == nil {
		return fallback, nil
	}

	return "", fmt.Errorf("No data.hson found in cwd or executable directory. Please specify a path to your HSON file using the --db or --database flag.")
}
