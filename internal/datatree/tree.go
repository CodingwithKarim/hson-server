package datatree

import (
	"errors"
	"fmt"
	"maps"
	"strings"
)

var ErrNotFound = errors.New("value not found in datatree")

func Lookup(appData any, urlPath string) (any, error) {
	// Split URL into separate segments | e.g: `/api/items/0` => [api, items, 0]
	urlParts := splitPath(urlPath)

	// Return full app data if URL path is root (`/`)
	if len(urlParts) == 0 {
		return appData, nil
	}

	// Traverse the app.Data to get the parent container & last segment for given URL path
	parentContainer, lastSegment, err := traverse(appData, urlParts)

	if err != nil {
		return nil, err
	}

	switch parent := parentContainer.(type) {
	case map[string]any:
		// If parent container is a map type, lookup map by using last segment as key
		result, ok := parent[lastSegment]

		if !ok {
			return nil, ErrNotFound
		}

		// Return result as lookup value
		return result, nil

	case []any:
		// Find target element in array by either ID prop or fallback to positional index
		result, _, err := findByKey(parent, lastSegment)

		if err != nil {
			return nil, ErrNotFound
		}

		// Return result as lookup value
		return result, nil

	default:
		// We only support maps/objects & arrays/slices for a parent container
		return nil, fmt.Errorf("cannot traverse into %T at %q", parent, lastSegment)
	}
}

func Set(root any, urlPath string, newVal any) error {
	// Split the URL path into segments e.g: `/api/books/1` => [api,books,1]
	urlParts := splitPath(urlPath)

	// Traverse the app.Data to get the parent container & last segment for given URL path
	parentContainer, lastSegment, err := traverse(root, urlParts)

	if err != nil {
		return err
	}

	// Set the new value in the parent container
	switch parent := parentContainer.(type) {
	case map[string]any:
		// If a map, assign new value to parent container
		parent[lastSegment] = newVal

		return nil

	case []any:
		// Find target index in array by either ID prop or fallback to positional index
		index, err := findIndex(parent, lastSegment)

		if err != nil {
			return err
		}

		// Assign the new value to correct location using index
		parent[index] = newVal
		return nil

	default:
		// We only support maps/objects & arrays/slices for a parent container
		return fmt.Errorf("cannot set on non-collection at %q", urlPath)
	}
}

func Delete(root any, urlPath string) error {
	// Split URL into separate segments | e.g: `/api/items/0` => [api, items, 0]
	urlParts := splitPath(urlPath)

	// Dont allow client to delete whole hson document by making DELETE request to root
	if len(urlParts) == 0 {
		return ErrNotFound
	}

	// Traverse the app.Data to get the parent container & last segment for given URL path
	parent, lastSegment, err := traverse(root, urlParts)

	if err != nil {
		return err
	}

	// Get the URL path for parent container
	parentPath := strings.Join(urlParts[:len(urlParts)-1], "/")

	switch parentContainer := parent.(type) {
	case map[string]any:
		// If the value of target element is an array, don't delete key + value, only clear list
		if _, ok := parentContainer[lastSegment].([]any); ok {
			parentContainer[lastSegment] = []any{}

			return nil
		}

		// Delete key + value pair from parent container
		delete(parentContainer, lastSegment)
		return nil

	case []any:
		// find the index for element we want to delete
		_, index, err := findByKey(parentContainer, lastSegment)

		if err != nil {
			return err
		}

		// Create a new slice by skipping index of requested deleted element
		newSlice := append(parentContainer[:index], parentContainer[index+1:]...)

		// Write the new slice back without the deleted element to app.data + persist
		return Set(root, parentPath, newSlice)

	default:
		// We only support maps/objects & arrays/slices for a parent container
		return fmt.Errorf("cannot delete non-collection at %q", urlPath)
	}
}

func BulkDelete(root any, urlPath string, filters map[string]string) error {
	// If no filters, perform a single delete (by ID or index).
	if len(filters) == 0 {
		return Delete(root, urlPath)
	}
	// Lookup the value at urlPath.
	raw, err := Lookup(root, urlPath)

	if err != nil {
		return err
	}

	// Make sure data we looked up is an arr for bulk deletion
	slice, ok := raw.([]any)

	if !ok {
		return fmt.Errorf("value at %q is not a slice", urlPath)
	}

	// Init a new arr to write back to app store
	kept := make([]any, 0, len(slice))

	// Loop through elements and only keep elements that don't match the filters
	// If they match the filters, we skip them and they inevitably get deleted
	for _, elem := range slice {
		if matchesFilters(elem, filters) {
			continue
		}

		kept = append(kept, elem)
	}

	// Persist the updated arr
	return Set(root, urlPath, kept)
}

func Patch(root any, urlPath string, patch map[string]any) error {
	// Split the URL path into segments e.g: /api/books/1 => [api,books,1]
	parts := splitPath(urlPath)

	// Get the parent container (map or array) and the final path segment  e.g: /api/book/1 => parent: /api/book & last: "1"
	parent, lastSegment, err := traverse(root, parts)

	if err != nil {
		return err
	}

	// Type check parent container to ensure it is a map type
	// Patch only valid for modifying objects / maps
	switch parentContainer := parent.(type) {
	case map[string]any:
		// Get the target value at the final segment
		data, exists := parentContainer[lastSegment]

		if !exists {
			return ErrNotFound
		}

		// Make sure target value is an object to allow valid patch
		targetObject, ok := data.(map[string]any)

		if !ok {
			return fmt.Errorf("cannot patch non-object at %q", urlPath)
		}

		// Copy patch / shallow merge into target object
		maps.Copy(targetObject, patch)

		return nil

	case []any:
		// Find target element in arr, ensuring it is an obj
		obj, _, err := findElem(parentContainer, lastSegment)

		if err != nil {
			return err
		}

		// Copy patch / shallow merge into target object
		maps.Copy(obj, patch)
		return nil

	default:
		return fmt.Errorf("cannot patch non-collection at %q", urlPath)
	}
}
