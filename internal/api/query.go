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

		if fmt.Sprint(record[key]) != values[0] {
			return false
		}
	}
	return true
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
