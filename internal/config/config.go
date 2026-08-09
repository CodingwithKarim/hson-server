package config

import (
	"flag"
	"fmt"

	"github.com/joho/godotenv"
)

type Config struct {
	DBPath       string
	ServerPort   string
	LogLevel     string
	LiveReload   bool
	IsLogVerbose bool
	Auth         *AuthConfig
}

func Load() (*Config, error) {
	// Load environment variables from a .env file if it exists
	_ = godotenv.Load()

	// Parse application flags and initialize the configuration struct
	cfg := parseAppFlags()

	// Resolve the path to the HSON database file
	resolvedPath, err := resolveDataFile(cfg.DBPath)

	if err != nil {
		return nil, err
	}

	// Update the configuration struct with the resolved database file path
	cfg.DBPath = resolvedPath

	// Load and validate the authentication configuration from auth.hjson
	if err := loadAuthData(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func parseAppFlags() *Config {
	cfg := &Config{}

	flag.StringVar(
		&cfg.DBPath,
		"db",
		envOrDefault("HSON_DB_PATH", "data.hson"),
		"path to your HSON database file",
	)

	flag.StringVar(
		&cfg.DBPath,
		"database",
		envOrDefault("HSON_DB_PATH", "data.hson"),
		"alias for --db",
	)

	flag.StringVar(
		&cfg.ServerPort,
		"port",
		envOrDefault("HSON_PORT", "3000"),
		"port the server will listen on",
	)

	flag.BoolVar(
		&cfg.LiveReload,
		"live-reload",
		envBoolOrDefault("HSON_LIVE_RELOAD", false),
		"watch HSON file and reload on external changes",
	)

	flag.StringVar(
		&cfg.LogLevel,
		"log-level",
		envOrDefault("HSON_LOG_LEVEL", "info"),
		"log level: debug, info, warn, error",
	)

	flag.BoolVar(
		&cfg.IsLogVerbose,
		"verbose",
		envBoolOrDefault("HSON_VERBOSE", false),
		"enable verbose logging (adds file and line number and extra fields)",
	)

	flag.Parse()

	return cfg
}

func resolveDataFile(dbPath string) (string, error) {
	if dbPath != "data.hson" {
		return dbPath, nil
	}

	path, err := resolveFile("data.hson")

	if err != nil {
		return "", fmt.Errorf("no data.hson found in cwd or executable directory; specify a path using --db or --database")
	}

	return path, nil
}
