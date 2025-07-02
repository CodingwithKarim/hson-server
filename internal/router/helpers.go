package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"hson-server/internal/datatree"
	"hson-server/internal/logger"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func parseQuery(qs url.Values) QueryOptions {
	opts := QueryOptions{Filters: make(map[string]string)}
	for key, values := range qs {
		v := values[0]
		switch key {
		case "sort":
			if v != "" {
				opts.Desc = v[0] == '-'
				if opts.Desc {
					v = v[1:]
				}
				opts.SortKey = v
			}
		case "limit":
			opts.Limit, _ = strconv.Atoi(v)
		case "page":
			if page, err := strconv.Atoi(v); err == nil && page > 0 && opts.Limit > 0 {
				opts.Offset = (page - 1) * opts.Limit
			}
		case "offset":
			opts.Offset, _ = strconv.Atoi(v)
		default:
			if v != "" {
				opts.Filters[key] = v
			}
		}
	}
	return opts
}

func validateJSONContentType(r *http.Request) error {
	if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
		return fmt.Errorf("Content-Type must be application/json")
	}

	return nil
}

func decodeJSONBody(request *http.Request, limit int64, dst any) error {
	// Limit the size of the request body
	request.Body = http.MaxBytesReader(nil, request.Body, limit)

	// Decode the JSON body into the destination variable
	return json.NewDecoder(request.Body).Decode(dst)
}

func ensureArray(resource any) ([]any, error) {
	switch v := resource.(type) {
	case []any:
		return v, nil
	case nil:
		logger.Warn(
			"Resource was nil; initializing empty array",
		)
		return []any{}, nil
	default:
		return nil, fmt.Errorf("POST only allowed on array endpoints")
	}
}

func handleStoreError(w http.ResponseWriter, r *http.Request, err error, context string) {
	if errors.Is(err, datatree.ErrNotFound) {
		logger.Error(
			context+": resource not found",
			"method", r.Method,
			"path", r.URL.Path,
			"query_params", r.URL.RawQuery,
			"err", datatree.ErrNotFound,
		)

		http.NotFound(w, r)
	} else {
		logger.Error(
			context+": internal error",
			"method", r.Method,
			"path", r.URL.Path,
			"query_params", r.URL.RawQuery,
			"err", err,
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func countItems(data any) int {
	switch v := data.(type) {
	case []any:
		return len(v)
	case map[string]any:
		return len(v)
	default:
		return 1
	}
}
