package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func resolveFile(filename string) (string, error) {
	if _, err := os.Stat(filename); err == nil {
		return filepath.Abs(filename)
	}

	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	fallback := filepath.Join(filepath.Dir(exePath), filename)

	if _, err := os.Stat(fallback); err == nil {
		return fallback, nil
	}

	return "", fmt.Errorf("%s not found in cwd or executable directory", filename)
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
