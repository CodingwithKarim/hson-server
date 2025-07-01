package datatree

import (
	"fmt"
	"path"
	"strconv"
	"strings"
)

// FlattenFilters converts multi-valued params to single-value map.
func FlattenFilters(vals map[string][]string) map[string]string {
	out := make(map[string]string, len(vals))
	for k, vs := range vals {
		if len(vs) > 0 {
			out[k] = vs[0]
		}
	}
	return out
}

func splitPath(urlPath string) []string {
	clean := path.Clean("/" + urlPath)
	trimmed := strings.Trim(clean, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

func traverse(root any, parts []string) (parent any, last string, err error) {
	// If path is empty, return out
	if len(parts) == 0 {
		return nil, "", ErrNotFound
	}

	// Init curr as root data (app.Data)
	curr := root

	// Iterate through (segments - 1) to hit the parent container
	for index, segment := range parts[:len(parts)-1] {
		prefix := strings.Join(parts[:index+1], "/")

		switch current := curr.(type) {
		case map[string]any:
			// If current value is map, traverse into map by key
			nxt, ok := current[segment]

			if !ok {
				return nil, "", fmt.Errorf("path not found: %q", prefix)
			}

			// Move to next element
			curr = nxt

		case []any:
			element, _, findErr := findByKey(current, segment)

			if findErr != nil {
				return nil, "", fmt.Errorf("invalid id/index %q at %q", segment, prefix)
			}

			// Move to next element
			curr = element

		default:
			// Cannot traverse deeper into other types, return out
			return nil, "", fmt.Errorf("cannot descend into non-collection at %q", prefix)
		}
	}

	// Return the parent container and the last path segment
	return curr, parts[len(parts)-1], nil
}

func findByKey(slice []any, key string) (elem any, idx int, err error) {
	key = strings.TrimSpace(key)

	// First, try matching an object's "id" field of any type.
	for i, el := range slice {
		m, ok := el.(map[string]any)
		if !ok {
			continue
		}
		raw, has := m["id"]

		if !has {
			continue
		}

		// 1) If the raw ID is a string, do a direct comparison.
		if s, ok := raw.(string); ok && s == key {
			return m, i, nil
		}

		// 3) If the raw ID is a float64 (the default for JSON numbers):
		if f, ok := raw.(float64); ok && strconv.Itoa(int(f)) == key {
			return m, i, nil
		}
	}

	// Fallback: treat key as a numeric index.
	if idx, err2 := strconv.Atoi(key); err2 == nil && idx >= 0 && idx < len(slice) {
		return slice[idx], idx, nil
	}

	return nil, -1, fmt.Errorf("no element with id or index %q", key)
}

// findElem returns the object (map[string]any) at key by id/index.
func findElem(slice []any, key string) (map[string]any, int, error) {
	el, idx, err := findByKey(slice, key)

	if err != nil {
		return nil, -1, err
	}
	m, ok := el.(map[string]any)

	if !ok {
		return nil, -1, fmt.Errorf("element at %q is not an object", key)
	}
	return m, idx, nil
}

// findIndex returns just the index of the element at key by id/index.
func findIndex(slice []any, key string) (int, error) {
	_, idx, err := findByKey(slice, key)
	return idx, err
}

func matchesFilters(elem any, filters map[string]string) bool {
	for key, want := range filters {
		switch obj := elem.(type) {
		case map[string]any:
			val, ok := obj[key]
			if !ok || fmt.Sprint(val) != want {
				// this filter didn’t match → don’t drop
				return false
			}
		default:
			// primitive slice: only "value" filter makes sense
			if key == "value" && fmt.Sprint(elem) == want {
				continue
			}
			return false
		}
	}
	// every filter matched → drop this element
	return true
}
