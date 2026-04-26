package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicopiov/zerapi/internal/loader"
	"github.com/nicopiov/zerapi/internal/store"
)

func newTestHandler() http.Handler {
	return newTestHandlerWithOptions(Options{})
}

func newTestHandlerWithOptions(options Options) http.Handler {
	data := store.New([]loader.Resource{
		{
			Name: "users",
			Records: []map[string]any{
				{"id": 1, "name": "Ada"},
				{"id": 2, "name": "Grace"},
			},
		},
	})

	return NewHandler(data, options)
}

func performRequest(handler http.Handler, method string, path string, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	return response
}

func decodeBody(t *testing.T, response *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	return body
}

func TestListResource(t *testing.T) {
	response := performRequest(newTestHandler(), http.MethodGet, "/users", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
}

func TestGetRecord(t *testing.T) {
	response := performRequest(newTestHandler(), http.MethodGet, "/users/1", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	body := decodeBody(t, response)
	if body["name"] != "Ada" {
		t.Fatalf("expected Ada, got %v", body["name"])
	}
}

func TestCreateRecord(t *testing.T) {
	response := performRequest(newTestHandler(), http.MethodPost, "/users", `{"name":"Linus"}`)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", response.Code)
	}

	body := decodeBody(t, response)
	if body["name"] != "Linus" {
		t.Fatalf("expected Linus, got %v", body["name"])
	}
}

func TestPatchRecord(t *testing.T) {
	response := performRequest(newTestHandler(), http.MethodPatch, "/users/1", `{"name":"Augusta"}`)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	body := decodeBody(t, response)
	if body["name"] != "Augusta" {
		t.Fatalf("expected Augusta, got %v", body["name"])
	}
}

func TestDeleteRecord(t *testing.T) {
	response := performRequest(newTestHandler(), http.MethodDelete, "/users/1", "")

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", response.Code)
	}
}

func TestMissingRecord(t *testing.T) {
	response := performRequest(newTestHandler(), http.MethodGet, "/users/999", "")

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.Code)
	}
}

func TestInvalidJSONBody(t *testing.T) {
	response := performRequest(newTestHandler(), http.MethodPost, "/users", `{`)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestWithLoggingWritesRequestLog(t *testing.T) {
	var output bytes.Buffer

	handler := WithLogging(newTestHandler(), &output)

	response := performRequest(handler, http.MethodGet, "/users/1", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	log := output.String()

	if !strings.Contains(log, "GET") {
		t.Fatalf("expected log to contain method, got %q", log)
	}

	if !strings.Contains(log, "/users/1") {
		t.Fatalf("expected log to contain path, got %q", log)
	}

	if !strings.Contains(log, "200") {
		t.Fatalf("expected log to contain status, got %q", log)
	}
}

func TestReadonlyAllowsReads(t *testing.T) {
	handler := newTestHandlerWithOptions(Options{Readonly: true})

	response := performRequest(handler, http.MethodGet, "/users/1", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
}

func TestReadonlyBlocksCreate(t *testing.T) {
	handler := newTestHandlerWithOptions(Options{Readonly: true})

	response := performRequest(handler, http.MethodPost, "/users", `{"name":"Linus"}`)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", response.Code)
	}
}

func TestReadonlyBlocksPatch(t *testing.T) {
	handler := newTestHandlerWithOptions(Options{Readonly: true})

	response := performRequest(handler, http.MethodPatch, "/users/1", `{"name":"Augusta"}`)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", response.Code)
	}
}

func TestReadonlyBlocksDelete(t *testing.T) {
	handler := newTestHandlerWithOptions(Options{Readonly: true})

	response := performRequest(handler, http.MethodDelete, "/users/1", "")

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", response.Code)
	}
}

func TestReadonlyBlocksReplace(t *testing.T) {
	handler := newTestHandlerWithOptions(Options{Readonly: true})

	response := performRequest(handler, http.MethodPut, "/users/1", `{"name":"Augusta"}`)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", response.Code)
	}
}

func TestFilterResourceByField(t *testing.T) {
	response := performRequest(newTestHandler(), http.MethodGet, "/users?name=Ada", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var body []map[string]any
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if len(body) != 1 {
		t.Fatalf("expected 1 record, got %d", len(body))
	}

	if body[0]["name"] != "Ada" {
		t.Fatalf("expected Ada, got %v", body[0]["name"])
	}
}

func TestFilterResourceNoMatches(t *testing.T) {
	response := performRequest(newTestHandler(), http.MethodGet, "/users?name=Missing", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var body []map[string]any
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if len(body) != 0 {
		t.Fatalf("expected 0 records, got %d", len(body))
	}
}

func TestFilterResourceByMultipleFields(t *testing.T) {
	response := performRequest(newTestHandler(), http.MethodGet, "/users?id=1&name=Ada", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var body []map[string]any
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if len(body) != 1 {
		t.Fatalf("expected 1 record, got %d", len(body))
	}

	if body[0]["name"] != "Ada" {
		t.Fatalf("expected Ada, got %v", body[0]["name"])
	}
}

func TestLimitResource(t *testing.T) {
	response := performRequest(newTestHandler(), http.MethodGet, "/users?_limit=1", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var body []map[string]any
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if len(body) != 1 {
		t.Fatalf("expected 1 record, got %d", len(body))
	}
}

func TestPageResource(t *testing.T) {
	response := performRequest(newTestHandler(), http.MethodGet, "/users?_page=2&_limit=1", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var body []map[string]any
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if len(body) != 1 {
		t.Fatalf("expected 1 record, got %d", len(body))
	}

	if body[0]["name"] != "Grace" {
		t.Fatalf("expected Grace, got %v", body[0]["name"])
	}
}

func TestInvalidLimitReturnsBadRequest(t *testing.T) {
	response := performRequest(newTestHandler(), http.MethodGet, "/users?_limit=abc", "")

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestInvalidPageReturnsBadRequest(t *testing.T) {
	response := performRequest(newTestHandler(), http.MethodGet, "/users?_page=abc&_limit=1", "")

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}
