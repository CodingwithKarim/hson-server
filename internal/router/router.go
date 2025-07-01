package router

import (
	"net/http"
	"net/url"
	"path"
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
	return addCORSAndNormalizeURL(handler)
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
