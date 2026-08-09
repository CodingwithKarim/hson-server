package router

import (
	"hson-server/internal/config"
	"hson-server/internal/logger"
	"net/http"
	"strings"
)

type AuthStrategy func(r *http.Request, authConfig *config.AuthConfig) bool

var authStrategies = map[string]AuthStrategy{
	"bearer":  authenticateBearer,
	"basic":   authenticateBasic,
	"api-key": authenticateAPIKey,
	"cookie":  authenticateCookie,
}

func authenticateBearer(r *http.Request, authConfig *config.AuthConfig) bool {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		logger.Error(
			"Authentication failed: authorization header is missing from request",
			"url_path", r.URL.Path,
		)

		return false
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		logger.Error(
			"Authentication failed: bearer prefix is missing from authorization header",
			"url_path", r.URL.Path,
		)

		return false
	}

	bearerTokenFromHeader := strings.TrimPrefix(authHeader, "Bearer ")

	if bearerTokenFromHeader == "" {
		logger.Error(
			"Authentication failed: bearer token is missing from authorization header",
			"url_path", r.URL.Path,
		)

		return false
	}

	if bearerTokenFromHeader != authConfig.Bearer.Token {
		logger.Error(
			"Authentication failed: invalid bearer token",
			"url_path", r.URL.Path,
		)

		return false
	}

	return true
}

func authenticateBasic(r *http.Request, authConfig *config.AuthConfig) bool {
	username, password, ok := r.BasicAuth()

	if !ok {
		logger.Error(
			"Authentication failed: failed to parse basic auth credentials",
			"url_path", r.URL.Path,
		)

		return false
	}

	if username != authConfig.Basic.Username || password != authConfig.Basic.Password {
		logger.Error(
			"Authentication failed: invalid basic auth credentials",
			"url_path", r.URL.Path,
		)

		return false
	}

	return true
}

func authenticateAPIKey(r *http.Request, authConfig *config.AuthConfig) bool {
	apiKeyHeader := r.Header.Get(authConfig.APIKey.Header)

	if apiKeyHeader == "" {
		logger.Error(
			"Authentication failed: API key is missing from request header",
			"url_path", r.URL.Path,
		)

		return false
	}

	if apiKeyHeader != authConfig.APIKey.Value {
		logger.Error(
			"Authentication failed: invalid API key",
			"url_path", r.URL.Path,
		)

		return false
	}

	return true
}

func authenticateCookie(r *http.Request, authConfig *config.AuthConfig) bool {
	cookie, err := r.Cookie(authConfig.Cookie.Name)

	if err != nil {
		logger.Error(
			"Authentication failed: cookie is missing from request",
			"url_path", r.URL.Path,
		)

		return false
	}

	if cookie.Value != authConfig.Cookie.Value {
		logger.Error(
			"Authentication failed: invalid cookie value",
			"url_path", r.URL.Path,
		)

		return false
	}

	return true
}
