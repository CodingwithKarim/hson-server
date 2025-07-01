package router

import (
	"encoding/json"
	"hson-server/internal/datatree"
	"hson-server/internal/logger"
	"net/http"
	"path"
	"strconv"
	"time"
)

func handleGetRequest(store HSONStore) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()

		// Get base URL path without query params e.g: /api/books?title=Harry Potter => /api/books
		path := request.URL.Path

		logger.Debug("Incoming GET request",
			"method", request.Method,
			"path", path,
			"query_params", request.URL.RawQuery,
			"user_agent", request.UserAgent(),
		)

		storeStart := time.Now()

		// Get data from the store based on the path
		data, readErr := store.Read(path)

		// Get the number of items
		dataCount := countItems(data)

		logger.Debug(
			"Store read/lookup result",
			"data", data,
			"path", path,
			"error", readErr,
			"item_count", dataCount,
			"store_duration", time.Since(storeStart),
		)

		if readErr != nil {
			handleStoreError(writer, request, readErr, "Lookup operation from store failed")
			return
		}

		// Get query params from URL e.g: /api/books?title=Harry Potter => { title: [Harry Potter] }
		queryParams := request.URL.Query()

		// Apply query params to filter results if provided
		filteredData := applyQuery(data, queryParams)

		filteredDataCount := countItems(filteredData)

		// Set Content type header to indiciate JSON response
		writer.Header().Set("Content-Type", "application/json")

		// Write the data into the response body as JSON
		status := http.StatusOK

		// JSON encode the filtered data as JSON and write it to the response body for client.
		if encodeErr := json.NewEncoder(writer).Encode(filteredData); encodeErr != nil {
			logger.Error(
				"Failed to encode JSON response",
				"filtered_count", filteredDataCount,
				"error", encodeErr,
			)

			status = http.StatusInternalServerError

			http.Error(writer, encodeErr.Error(), status)
		}

		logger.Info("GET request completed ✅",
			"path", path,
			"query_params", request.URL.RawQuery,
			"status", status,
			"raw_count", dataCount,
			"filtered_count", filteredDataCount,
			"request_duration", time.Since(start),
		)
	}
}

func handlePostRequest(store HSONStore) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()

		logger.Debug("Incoming POST request",
			"method", request.Method,
			"path", request.URL.Path,
			"user_agent", request.UserAgent(),
		)

		// Ensure proper content type header
		if err := validateJSONContentType(request); err != nil {
			logger.Warn("Unsupported media type", "path", request.URL.Path, "err", err)
			http.Error(writer, err.Error(), http.StatusUnsupportedMediaType)
			return
		}

		// Init variable for new item
		var newItem any

		// Decode JSON request body into newItem variable
		if err := decodeJSONBody(request, 1<<20, &newItem); err != nil {
			logger.Error("Invalid JSON body", "path", request.URL.Path, "err", err)
			http.Error(writer, "Invalid JSON", http.StatusBadRequest)
			return
		}

		storeStart := time.Now()

		// Read the existing resource/value at the given path
		existingResource, readErr := store.Read(request.URL.Path)

		existingCount := countItems(existingResource)

		logger.Debug(
			"Store read/lookup result",
			"data", existingResource,
			"path", request.URL.Path,
			"error", readErr,
			"initial_count", existingCount,
			"store_duration", time.Since(storeStart),
		)

		if readErr != nil && readErr != datatree.ErrNotFound {
			handleStoreError(writer, request, readErr, "Data lookup failed")
			return
		}

		// Ensure the value of existingResource is an array
		// If null, init a new array
		arr, err := ensureArray(existingResource)

		if err != nil {
			logger.Error("Cannot POST to non-array endpoint, try PUT instead", "path", request.URL.Path, "err", err)
			writer.Header().Set("Allow", "GET,PUT,DELETE")
			http.Error(writer, err.Error(), http.StatusMethodNotAllowed)
			return
		}

		// Append the new item to the array
		arr = append(arr, newItem)

		writeStart := time.Now()

		// Write the updated array back to the store at the URL path
		if err := store.Write(request.URL.Path, arr); err != nil {
			handleStoreError(writer, request, err, "Writing operation from store failed")
			return
		}

		logger.Debug(
			"Appended new item to array and persisted change",
			"path", request.URL.Path,
			"new_count", len(arr),
			"write_duration", time.Since(writeStart),
		)

		// Get the index of the newly appended item
		index := len(arr) - 1

		// Construct location string using url path and new item index
		location := path.Join(request.URL.Path, strconv.Itoa(index))

		// Set Content type header to indiciate JSON response
		writer.Header().Set("Content-Type", "application/json")

		// Set the Location header to point to the newly created item's path
		writer.Header().Set("Location", location)

		// Respond with 201 Created status
		writer.WriteHeader(http.StatusCreated)

		// Write the array to response body as JSON
		if err := json.NewEncoder(writer).Encode(arr); err != nil {
			logger.Error(
				"Failed to encode response JSON",
				"path", request.URL.Path,
				"error", err,
			)
		}

		logger.Info("POST request completed ✅",
			"path", request.URL.Path,
			"location", location,
			"item_count", len(arr),
			"new_item", newItem,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}

func handlePutRequest(store HSONStore) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()

		logger.Debug("Incoming PUT request",
			"method", request.Method,
			"path", request.URL.Path,
			"user_agent", request.UserAgent(),
		)

		// Ensure proper content type header
		if err := validateJSONContentType(request); err != nil {
			logger.Error("Unsupported media type", "path", request.URL.Path, "err", err)
			http.Error(writer, err.Error(), http.StatusUnsupportedMediaType)
			return
		}

		// Init variable for value
		var newValue any

		// Decode JSON request body into newValue variable
		if err := decodeJSONBody(request, 1<<20, &newValue); err != nil {
			logger.Error("Invalid JSON body", "path", request.URL.Path, "err", err)
			http.Error(writer, "Invalid JSON", http.StatusBadRequest)
			return
		}

		writeStart := time.Now()

		// Write the updated value back to the store at the URL path
		if err := store.Write(request.URL.Path, newValue); err != nil {
			handleStoreError(writer, request, err, "Writing operation from store failed")
			return
		}

		logger.Debug(
			"Successfully updated value in store",
			"path", request.URL.Path,
			"write_duration", time.Since(writeStart),
			"value", newValue,
		)

		// Respond with 204 No Conent
		writer.WriteHeader(http.StatusNoContent)

		logger.Info("PUT request completed ✅",
			"path", request.URL.Path,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}

func handlePatchRequest(store HSONStore) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()

		logger.Debug("Incoming PATCH request",
			"method", request.Method,
			"path", request.URL.Path,
			"user_agent", request.UserAgent(),
		)

		// Ensure proper content type header
		if err := validateJSONContentType(request); err != nil {
			logger.Error("Unsupported media type", "path", request.URL.Path, "err", err)
			http.Error(writer, err.Error(), http.StatusUnsupportedMediaType)
			return
		}

		// Init variable for patch
		var patch map[string]any

		// Decode JSON request body into patch variable
		if err := decodeJSONBody(request, 1<<20, &patch); err != nil {
			logger.Error("Invalid JSON body", "path", request.URL.Path, "err", err)
			http.Error(writer, "invalid JSON", http.StatusBadRequest)
			return
		}

		logger.Debug("Decoded patch payload", "path", request.URL.Path, "patch", patch)

		patchStart := time.Now()

		// Apply patch to the value at the given path
		if err := store.Patch(request.URL.Path, patch); err != nil {
			handleStoreError(writer, request, err, "Patch operation from store failed")
			return
		}

		logger.Debug("Patch applied to store",
			"path", request.URL.Path,
			"patch_duration_ms", time.Since(patchStart).Milliseconds(),
		)

		// Respond with 204 No Conent
		writer.WriteHeader(http.StatusNoContent)

		logger.Info("PATCH request completed ✅",
			"path", request.URL.Path,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}

func handleDeleteRequest(store HSONStore) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()

		// Get base URL path without query params e.g: /api/books?title=Harry Potter => /api/books
		path := request.URL.Path

		logger.Debug("Incoming DELETE request",
			"method", request.Method,
			"path", path,
			"query_params", request.URL.RawQuery,
			"user_agent", request.UserAgent(),
		)

		storeStart := time.Now()

		// Delete resource at Path + any potential filters & persist change
		err := store.Delete(path, request.URL.Query())

		logger.Debug("Store delete result",
			"path", path,
			"error", err,
			"delete_duration", time.Since(storeStart),
		)

		if err != nil {
			handleStoreError(writer, request, err, "DELETE operation from store failed")
			return
		}

		// Respond with 204 to client No Conent
		writer.WriteHeader(http.StatusNoContent)

		logger.Info("DELETE request completed ✅",
			"path", path,
			"status", http.StatusNoContent,
			"request_duration", time.Since(start),
		)
	}
}
