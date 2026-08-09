package router

import (
	"hson-server/internal/config"
	"hson-server/internal/logger"
	"net/http"
	"path"
	"time"
)

func addAuth(next http.Handler, authConfig *config.AuthConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !authConfig.Enabled || !isProtectedRoute(r.URL.Path, authConfig.ProtectedRoutes) {
			next.ServeHTTP(w, r)
			return
		}

		authStrategy, ok := authStrategies[authConfig.Type]

		if !ok || !authStrategy(r, authConfig) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func addCORSAndNormalizeURL(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Normalize URL path
		r.URL.Path = path.Clean("/" + r.URL.Path)

		// Set CORS headers for the response
		setCORSHeaders(w, r)

		// Handle browser CORS preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

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
