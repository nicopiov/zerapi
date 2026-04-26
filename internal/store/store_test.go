package store

import (
	"testing"

	"github.com/nicopiov/zerapi/internal/loader"
)

func newTestStore() *Store {
	return New([]loader.Resource{
		{
			Name: "users",
			Records: []map[string]any{
				{"id": 1, "name": "Ada"},
				{"id": 2, "name": "Grace"},
			},
		},
	})
}

func TestList(t *testing.T) {
	store := newTestStore()

	records, ok := store.List("users")
	if !ok {
		t.Fatal("expected users resource")
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
}

func TestGet(t *testing.T) {
	store := newTestStore()

	record, ok := store.Get("users", "1")
	if !ok {
		t.Fatal("expected record")
	}

	if record["name"] != "Ada" {
		t.Fatalf("expected Ada, got %v", record["name"])
	}
}

func TestCreateAssignsID(t *testing.T) {
	store := newTestStore()

	record, ok := store.Create("users", map[string]any{"name": "Linus"})
	if !ok {
		t.Fatal("expected create to succeed")
	}

	if record["id"] != 3 {
		t.Fatalf("expected id 3, got %v", record["id"])
	}
}

func TestReplace(t *testing.T) {
	store := newTestStore()

	record, ok := store.Replace("users", "1", map[string]any{"name": "Augusta"})
	if !ok {
		t.Fatal("expected replace to succeed")
	}

	if record["id"] != 1 {
		t.Fatalf("expected id 1, got %v", record["id"])
	}

	if record["name"] != "Augusta" {
		t.Fatalf("expected Augusta, got %v", record["name"])
	}
}

func TestPatch(t *testing.T) {
	store := newTestStore()

	record, ok := store.Patch("users", "1", map[string]any{"name": "Augusta"})
	if !ok {
		t.Fatal("expected patch to succeed")
	}

	if record["name"] != "Augusta" {
		t.Fatalf("expected Augusta, got %v", record["name"])
	}
}

func TestDelete(t *testing.T) {
	store := newTestStore()

	if !store.Delete("users", "1") {
		t.Fatal("expected delete to succeed")
	}

	if _, ok := store.Get("users", "1"); ok {
		t.Fatal("expected record to be deleted")
	}
}

func TestMissingResource(t *testing.T) {
	store := newTestStore()

	if _, ok := store.List("posts"); ok {
		t.Fatal("expected missing resource")
	}
}

func TestPatchPreservesExistingFields(t *testing.T) {
	store := newTestStore()

	record, ok := store.Patch("users", "1", map[string]any{"name": "Augusta"})
	if !ok {
		t.Fatal("expected patch to succeed")
	}

	if record["id"] != 1 {
		t.Fatalf("expected id 1, got %v", record["id"])
	}
}

func TestReturnedRecordsAreCopies(t *testing.T) {
	store := newTestStore()

	record, ok := store.Get("users", "1")
	if !ok {
		t.Fatal("expected record")
	}

	record["name"] = "Changed"

	again, ok := store.Get("users", "1")
	if !ok {
		t.Fatal("expected record")
	}

	if again["name"] != "Ada" {
		t.Fatalf("expected internal record to stay Ada, got %v", again["name"])
	}
}

func TestReloadReplacesResources(t *testing.T) {
	store := newTestStore()

	store.Reload([]loader.Resource{
		{
			Name: "posts",
			Records: []map[string]any{
				{"id": 1, "title": "Hello"},
			},
		},
	})

	if _, ok := store.List("users"); ok {
		t.Fatal("expected users resource to be removed")
	}

	posts, ok := store.List("posts")
	if !ok {
		t.Fatal("expected posts resource")
	}

	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
}
