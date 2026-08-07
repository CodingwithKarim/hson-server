package router

import (
	"hson-server/internal/logger"
	"net/http"
	"path"
	"time"
)

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
