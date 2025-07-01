package router

import (
	"fmt"
	"hson-server/internal/logger"
	"net/url"
	"sort"
	"strconv"
	"time"
)

type QueryOptions struct {
	Filters map[string]string
	SortKey string
	Desc    bool
	Offset  int
	Limit   int
}

func applyQuery(rawData any, qs url.Values) any {
	// Attempt to parse raw data to an array since query params only supported on arrays
	arr, ok := rawData.([]any)

	// If data isn't an array or empty array, return raw data
	if !ok || len(qs) == 0 {
		return rawData
	}

	// Get all necessary filter options from url values
	filterOptions := parseQuery(qs)

	start := time.Now()

	// Return a new data slice with filters applied
	data := pipeline(arr, filterOptions)

	logger.Debug("Applied query filters",
		"filters", qs,
		"items_count", countItems(data),
		"filter_duration", time.Since(start),
	)

	// Return new data with filters applied
	return data
}

// pipeline applies filter → sort → paginate in order.
func pipeline(arr []any, opts QueryOptions) []any {
	if len(opts.Filters) > 0 {
		arr = filterArray(arr, opts.Filters)
	}
	if opts.SortKey != "" {
		sortArray(arr, opts.SortKey, opts.Desc)
	}
	return paginateArray(arr, opts.Offset, opts.Limit)
}

// filterArray retains only items matching all filter key-values.
func filterArray(arr []any, filters map[string]string) []any {
	out := make([]any, 0, len(arr))
	for _, item := range arr {
		if obj, ok := item.(map[string]any); ok {
			match := true
			for k, want := range filters {
				val, exists := obj[k]

				if !exists {
					match = false
					break
				}

				switch v := val.(type) {
				case float64:
					w, err := strconv.ParseFloat(want, 64)
					if err != nil || v != w {
						match = false
					}
				default:
					if fmt.Sprint(v) != want {
						match = false
					}
				}

				if !match {
					break
				}
			}
			if match {
				out = append(out, item)
			}
		} else if want, ok := filters["value"]; ok {
			if fmt.Sprint(item) == want {
				out = append(out, item)
			}
		}
	}
	return out
}

// sortArray orders the slice by key, ascending or descending.
func sortArray(arr []any, key string, desc bool) {
	sort.SliceStable(arr, func(i, j int) bool {
		ai, aok := arr[i].(map[string]any)
		bi, bok := arr[j].(map[string]any)
		if !aok || !bok {
			return false
		}

		aVal, aok := ai[key]
		bVal, bok := bi[key]
		if !aok || !bok {
			return false
		}

		switch av := aVal.(type) {
		case float64:
			bv, ok := bVal.(float64)
			if !ok {
				return !desc
			}
			if desc {
				return bv < av
			}
			return av < bv

		case string:
			bv, ok := bVal.(string)
			if !ok {
				return desc
			}
			if desc {
				return bv < av
			}
			return av < bv

		default:
			return false
		}
	})
}

// paginateArray slices the array according to offset and limit.
func paginateArray(arr []any, offset, limit int) []any {
	n := len(arr)
	if offset < 0 {
		offset = 0
	} else if offset >= n {
		return []any{}
	}
	end := n
	if limit > 0 && offset+limit < n {
		end = offset + limit
	}
	return arr[offset:end]
}
