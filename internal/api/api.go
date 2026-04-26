package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/nicopiov/zerapi/internal/store"
)

type Handler struct {
	store    *store.Store
	readonly bool
}

type Options struct {
	Readonly bool
}

func NewHandler(store *store.Store, options Options) http.Handler {
	return &Handler{store: store, readonly: options.Readonly}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := pathParts(r.URL.Path)

	if len(parts) == 0 {
		writeJSON(w, http.StatusOK, map[string]string{
			"name":   "zerapi",
			"status": "running",
		})
		return
	}

	if len(parts) > 2 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	resource := parts[0]

	if len(parts) == 1 {
		h.handleCollection(w, r, resource)
		return
	}

	h.handleRecord(w, r, resource, parts[1])
}

func (h *Handler) handleCollection(w http.ResponseWriter, r *http.Request, resource string) {
	switch r.Method {
	case http.MethodGet:
		records, ok := h.store.List(resource)
		if !ok {
			writeError(w, http.StatusNotFound, "resource not found")
			return
		}
		records = applyFilters(records, r.URL.Query())
		applySorting(records, r.URL.Query())

		records, ok = applyPagination(w, records, r.URL.Query())
		if !ok {
			return
		}

		writeJSON(w, http.StatusOK, records)

	case http.MethodPost:
		if h.readonly {
			writeError(w, http.StatusForbidden, "readonly mode")
			return
		}

		record, ok := readRecord(w, r)
		if !ok {
			return
		}

		created, ok := h.store.Create(resource, record)
		if !ok {
			writeError(w, http.StatusNotFound, "resource not found")
			return
		}
		writeJSON(w, http.StatusCreated, created)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleRecord(w http.ResponseWriter, r *http.Request, resource string, id string) {
	switch r.Method {
	case http.MethodGet:
		record, ok := h.store.Get(resource, id)
		if !ok {
			writeError(w, http.StatusNotFound, "record not found")
			return
		}
		writeJSON(w, http.StatusOK, record)

	case http.MethodPut:
		if h.readonly {
			writeError(w, http.StatusForbidden, "readonly mode")
			return
		}

		record, ok := readRecord(w, r)
		if !ok {
			return
		}

		replaced, ok := h.store.Replace(resource, id, record)
		if !ok {
			writeError(w, http.StatusNotFound, "record not found")
			return
		}
		writeJSON(w, http.StatusOK, replaced)

	case http.MethodPatch:
		if h.readonly {
			writeError(w, http.StatusForbidden, "readonly mode")
			return
		}

		patch, ok := readRecord(w, r)
		if !ok {
			return
		}

		patched, ok := h.store.Patch(resource, id, patch)
		if !ok {
			writeError(w, http.StatusNotFound, "record not found")
			return
		}
		writeJSON(w, http.StatusOK, patched)

	case http.MethodDelete:
		if h.readonly {
			writeError(w, http.StatusForbidden, "readonly mode")
			return
		}

		if !h.store.Delete(resource, id) {
			writeError(w, http.StatusNotFound, "record not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func readRecord(w http.ResponseWriter, r *http.Request) (map[string]any, bool) {
	var record map[string]any
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return nil, false
	}

	if record == nil {
		writeError(w, http.StatusBadRequest, "json body must be an object")
		return nil, false
	}
	return record, true
}

func pathParts(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}

	return strings.Split(trimmed, "/")
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

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
