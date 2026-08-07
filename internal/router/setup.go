package router

import (
	"hson-server/internal/config"
	"net/http"
	"net/url"
)

// HSONStore defines operations for reading/writing HSON data
// Inferface is implemented in app package
type HSONStore interface {
	Read(path string) (any, error)
	Write(path string, newVal any) error
	Delete(path string, values url.Values) error
	Patch(path string, patchData map[string]any) error
}

func NewHTTPHandler(store HSONStore, authConfig *config.AuthConfig) http.Handler {
	// Assemble a HTTP multiplexer (router)
	handler := http.NewServeMux()

	// Register a dispatcher function at the root path
	handler.HandleFunc("/", handlerDispatcher(store))

	// Return the configured router with CORS, authentication, and delay middleware applied
	return addCORSAndNormalizeURL(
		addAuth(
			addDelay(handler),
			authConfig,
		),
	)
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
