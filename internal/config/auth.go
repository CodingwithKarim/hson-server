package config

import (
	"fmt"
	"os"
	"slices"

	"github.com/hjson/hjson-go/v4"
)

type AuthConfig struct {
	Enabled         bool             `json:"enabled"`
	Type            string           `json:"type"`
	ProtectedRoutes []string         `json:"protectedRoutes"`
	Bearer          BearerAuthConfig `json:"bearer"`
	Basic           BasicAuthConfig  `json:"basic"`
	APIKey          APIKeyAuthConfig `json:"apiKey"`
	Cookie          CookieAuthConfig `json:"cookie"`
}

type BearerAuthConfig struct {
	Token string `json:"token"`
}

type BasicAuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type APIKeyAuthConfig struct {
	Header string `json:"header"`
	Value  string `json:"value"`
}

type CookieAuthConfig struct {
	IssueRoute string `json:"issueRoute"`
	Name       string `json:"name"`
	Value      string `json:"value"`
}

var authTypes []string = []string{"api-key", "basic", "bearer", "cookie"}

func loadAuthData(cfg *Config) error {
	// Resolve the path to the auth.hjson file
	authPath, err := resolveAuthFile()

	if err != nil {
		return err
	}

	// Read the contents of the auth.hjson file
	raw, err := os.ReadFile(authPath)

	if err != nil {
		return err
	}

	var data AuthConfig

	// Unmarshal the raw HJSON data into the AuthConfig struct
	if err := hjson.Unmarshal(raw, &data); err != nil {
		return err
	}

	// Validate the loaded authentication configuration
	if err := validateAuthConfig(&data); err != nil {
		return err
	}

	// Update the configuration struct with the loaded authentication data
	cfg.Auth = &data

	return nil
}

func validateAuthConfig(authConfig *AuthConfig) error {
	// If authentication is not enabled, no further validation is required
	if !authConfig.Enabled {
		return nil
	}

	// Check if the specified authentication type is supported
	if !slices.Contains(authTypes, authConfig.Type) {
		return fmt.Errorf("unsupported authentication type: %s", authConfig.Type)
	}

	// Perform type-specific validation for the configured authentication type
	switch authConfig.Type {
	case "bearer":
		if authConfig.Bearer.Token == "" {
			return fmt.Errorf("bearer authentication requires a bearer token")
		}

	case "basic":
		if authConfig.Basic.Username == "" || authConfig.Basic.Password == "" {
			return fmt.Errorf("basic authentication requires username and password")
		}

	case "api-key":
		if authConfig.APIKey.Header == "" || authConfig.APIKey.Value == "" {
			return fmt.Errorf("api-key authentication requires header and value")
		}

	case "cookie":
		if authConfig.Cookie.Name == "" || authConfig.Cookie.Value == "" || authConfig.Cookie.IssueRoute == "" {
			return fmt.Errorf("cookie authentication requires name, value, and issue route")
		}
	}

	return nil
}

func resolveAuthFile() (string, error) {
	return resolveFile("auth.hjson")
}
