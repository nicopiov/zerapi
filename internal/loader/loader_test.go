package loader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTopLevelArray(t *testing.T) {
	path := writeTempFile(t, "users.json", `[
		{"id": 1, "name": "Ada"},
		{"id": 2, "name": "Grace"}
	]`)

	result, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(result.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(result.Resources))
	}

	resource := result.Resources[0]

	if resource.Name != "users" {
		t.Fatalf("expected resource name users, got %q", resource.Name)
	}

	if len(resource.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(resource.Records))
	}
}

func TestLoadTopLevelObject(t *testing.T) {
	path := writeTempFile(t, "db.json", `{
		"users": [
			{"id": 1, "name": "Ada"}
		],
		"posts": [
			{"id": 1, "title": "Hello"}
		]
	}`)

	result, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	resources := map[string]int{}
	for _, resource := range result.Resources {
		resources[resource.Name] = len(resource.Records)
	}

	if resources["users"] != 1 {
		t.Fatalf("expected users resource with 1 record, got %d", resources["users"])
	}

	if resources["posts"] != 1 {
		t.Fatalf("expected posts resource with 1 record, got %d", resources["posts"])
	}
}

func TestLoadRejectsInvalidJSON(t *testing.T) {
	path := writeTempFile(t, "bad.json", `{`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadRejectsNonObjectRecords(t *testing.T) {
	path := writeTempFile(t, "users.json", `["Ada"]`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadRejectsObjectPropertiesThatAreNotArrays(t *testing.T) {
	path := writeTempFile(t, "db.json", `{
		"users": {"id": 1}
	}`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func writeTempFile(t *testing.T, name string, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	return path
}
