package router

import (
	"hson-server/internal/logger"
	"net/http"
	"net/url"
	"path"
	"time"
)

// HSONStore defines operations for reading/writing HSON data
// Inferface is implemented in app package
type HSONStore interface {
	Read(path string) (any, error)
	Write(path string, newVal any) error
	Delete(path string, values url.Values) error
	Patch(path string, patchData map[string]any) error
}

func NewHTTPHandler(store HSONStore) http.Handler {
	// Assemble a HTTP multiplexer (router)
	handler := http.NewServeMux()

	// Register a dispatcher function at the root path
	handler.HandleFunc("/", handlerDispatcher(store))

	// Return the configured router
	return addCORSAndNormalizeURL(addDelay(handler))
}

// Depending on the HTTP verb, we will dispatch its equivalent handler function
// If HTTP verb is not supported, set the allow header and return an error to client
func handlerDispatcher(store HSONStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetRequest(store)(w, r)
		case http.MethodPost:
			handlePostRequest(store)(w, r)
		case http.MethodPut:
			handlePutRequest(store)(w, r)
		case http.MethodPatch:
			handlePatchRequest(store)(w, r)
		case http.MethodDelete:
			handleDeleteRequest(store)(w, r)
		default:
			w.Header().Set("Allow", "GET,POST,PUT,PATCH,DELETE")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
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
				"value", delayString,
				"err", err,
			)
			next.ServeHTTP(w, r)
			return
		}

		if duration <= 0 {
			logger.Warn("Invalid delay duration",
				"value", delayString,
			)
			next.ServeHTTP(w, r)
			return
		}

		if duration > time.Minute {
			logger.Warn("Delay duration too long",
				"requested_duration", duration,
				"adjusted_to", time.Minute,
			)
			duration = time.Minute
		}

		select {
		case <-time.After(duration):
			next.ServeHTTP(w, r)
		case <-r.Context().Done():
			logger.Info("Request cancelled by client during delay",
				"requested_duration", duration,
			)
			return
		}
	})
}
