package api

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

func applyFilters(records []map[string]any, query url.Values) []map[string]any {
	filtered := make([]map[string]any, 0, len(records))

	for _, record := range records {
		if matchesFilters(record, query) {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

func matchesFilters(record map[string]any, query url.Values) bool {
	for key, values := range query {
		if isReservedQueryParam(key) {
			continue
		}

		if len(values) == 0 {
			continue
		}

		if !matchesFilter(record, key, values[0]) {
			return false
		}

	}
	return true
}

func matchesFilter(record map[string]any, key string, expected string) bool {
	field, operator := splitFilterKey(key)
	actual, ok := record[field]
	if !ok {
		return false
	}

	switch operator {
	case "exact":
		return fmt.Sprint(actual) == expected
	case "like":
		return strings.Contains(
			strings.ToLower(fmt.Sprint(actual)),
			strings.ToLower(expected),
		)
	case "gte":
		return compareNumber(actual, expected, func(left float64, right float64) bool {
			return left >= right
		})
	case "lte":
		return compareNumber(actual, expected, func(left float64, right float64) bool {
			return left <= right
		})
	default:
		return false
	}
}

func splitFilterKey(key string) (string, string) {
	for _, suffix := range []string{"_like", "_gte", "_lte"} {
		if strings.HasSuffix(key, suffix) {
			return strings.TrimSuffix(key, suffix), strings.TrimPrefix(suffix, "_")
		}
	}

	return key, "exact"
}

func compareNumber(actual any, expected string, compare func(float64, float64) bool) bool {
	left, ok := numberValue(actual)
	if !ok {
		return false
	}

	right, err := strconv.ParseFloat(expected, 64)
	if err != nil {
		return false
	}

	return compare(left, right)
}

func numberValue(value any) (float64, bool) {
	switch typed := value.(type) {
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case float64:
		return typed, true
	case string:
		parsed, err := strconv.ParseFloat(typed, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func isReservedQueryParam(key string) bool {
	switch key {
	case "_page", "_limit", "_sort":
		return true
	default:
		return false
	}
}

func applyPagination(w http.ResponseWriter, records []map[string]any, query url.Values) ([]map[string]any, bool) {
	limitValue := query.Get("_limit")
	if limitValue == "" {
		return records, true
	}

	limit, err := strconv.Atoi(limitValue)
	if err != nil || limit < 1 {
		writeError(w, http.StatusBadRequest, "_limit must be a positive integer")
		return nil, false
	}

	page := 1
	pageValue := query.Get("_page")
	if pageValue != "" {
		parsedPage, err := strconv.Atoi(pageValue)
		if err != nil || parsedPage < 1 {
			writeError(w, http.StatusBadRequest, "_page must be a positive integer")
			return nil, false
		}
		page = parsedPage
	}

	start := (page - 1) * limit
	if start >= len(records) {
		return []map[string]any{}, true
	}

	end := start + limit
	if end > len(records) {
		end = len(records)
	}

	return records[start:end], true
}

func applySorting(records []map[string]any, query url.Values) {
	sortValue := query.Get("_sort")
	if sortValue == "" {
		return
	}

	descending := strings.HasPrefix(sortValue, "-")
	field := strings.TrimPrefix(sortValue, "-")

	if field == "" {
		return
	}

	sort.SliceStable(records, func(i, j int) bool {
		left := fmt.Sprint(records[i][field])
		right := fmt.Sprint(records[j][field])

		if descending {
			return left > right
		}

		return left < right
	})
}
