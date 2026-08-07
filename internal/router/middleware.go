package router

import (
	"hson-server/internal/config"
	"hson-server/internal/logger"
	"net/http"
	"path"
	"strings"
	"time"
)

func addAuth(next http.Handler, authConfig *config.AuthConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !authConfig.Enabled || !isProtectedRoute(r.URL.Path, authConfig.ProtectedRoutes) {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")

		if !strings.HasPrefix(authHeader, "Bearer ") {
			logger.Error("Bearer prefix is missing from authorization header", "auth_header", authHeader, "url_path", r.URL.Path)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		apiKeyFromHeader := strings.TrimPrefix(authHeader, "Bearer ")

		if apiKeyFromHeader == "" {
			logger.Error("Bearer api key is missing from authorization header", "auth_header", authHeader, "url_path", r.URL.Path)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if apiKeyFromHeader != authConfig.APIKey {
			logger.Error("Invalid bearer api key", "auth_header", authHeader, "url_path", r.URL.Path, "api_key_from_header", apiKeyFromHeader, "expected_api_key", authConfig.APIKey)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func addCORSAndNormalizeURL(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Clean URL path from request
		r.URL.Path = path.Clean("/" + r.URL.Path)

		// Set necessary CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

		// Handle potential OPTIONS requests from browsers
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func addDelay(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		delayString := r.URL.Query().Get("delay")

		if delayString == "" {
			next.ServeHTTP(w, r)
			return
		}

		duration, err := time.ParseDuration(delayString)

		if err != nil {
			logger.Error("Failed to parse delay duration",
				"delay_string", delayString,
				"err", err,
			)
			next.ServeHTTP(w, r)
			return
		}

		if duration <= 0 {
			logger.Warn("Invalid delay duration",
				"delay_string", delayString,
				"delay", duration,
			)
			next.ServeHTTP(w, r)
			return
		}

		if duration > time.Minute {
			logger.Warn("Delay duration too long",
				"delay_string", delayString,
				"delay", duration,
				"adjusted_to", time.Minute,
			)
			duration = time.Minute
		}

		select {
		case <-time.After(duration):
			next.ServeHTTP(w, r)
		case <-r.Context().Done():
			logger.Info("Request cancelled by client during delay",
				"delay_string", delayString,
				"delay", duration,
			)
			return
		}
	})
}
