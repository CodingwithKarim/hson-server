package router

import (
	"hson-server/internal/config"
	"net/http"
	"net/url"
)

// HSONStore interface defines operations for reading/writing HSON data which is implemented in the app package
type HSONStore interface {
	Read(path string) (any, error)
	Write(path string, newVal any) error
	Delete(path string, values url.Values) error
	Patch(path string, patchData map[string]any) error
}

func NewHTTPHandler(store HSONStore, authConfig *config.AuthConfig) http.Handler {
	// Setup the HTTP routes and return the configured handler
	handler := setupRoutes(store, authConfig)

	// Apply middleware for CORS, authentication, and request delay before returning the handler
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

func setupRoutes(store HSONStore, authConfig *config.AuthConfig) *http.ServeMux {
	handler := http.NewServeMux()

	// Register the cookie issue route if cookie authentication is enabled
	if authConfig.Enabled && authConfig.Type == "cookie" {
		handler.HandleFunc(
			authConfig.Cookie.IssueRoute,
			handleIssueCookie(authConfig),
		)
	}

	// Register the main handler dispatcher at the root path for all other requests
	handler.HandleFunc("/", handlerDispatcher(store))

	return handler
}
