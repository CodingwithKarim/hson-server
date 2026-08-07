package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/hjson/hjson-go/v4"
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

type AuthConfig struct {
	Enabled         bool     `json:"enabled"`
	APIKey          string   `json:"apiKey"`
	ProtectedRoutes []string `json:"protectedRoutes"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := parseAppFlags()

	resolvedPath, err := resolveDataFile(cfg.DBPath)

	if err != nil {
		return nil, err
	}

	cfg.DBPath = resolvedPath

	if err := loadAuthData(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadAuthData(cfg *Config) error {
	raw, err := os.ReadFile("auth.hson")

	if err != nil {
		return err
	}

	var data AuthConfig

	if err := hjson.Unmarshal(raw, &data); err != nil {
		return err
	}

	cfg.Auth = &data

	return nil
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

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}
	return value
}

func envBoolOrDefault(key string, fallback bool) bool {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}
