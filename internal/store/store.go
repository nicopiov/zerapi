package store

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/nicopiov/zerapi/internal/loader"
)

type Store struct {
	mu        sync.RWMutex
	resources map[string]*Resource
}

type Resource struct {
	Name    string
	Records []map[string]any
	nextID  int
}

func New(resources []loader.Resource) *Store {
	store := &Store{
		resources: make(map[string]*Resource, len(resources)),
	}

	for _, resource := range resources {
		records := copyRecords(resource.Records)

		nextID := 1
		for _, record := range records {
			id, ok := intID(record["id"])
			if ok && id >= nextID {
				nextID = id + 1
			}
		}

		store.resources[resource.Name] = &Resource{
			Name:    resource.Name,
			Records: records,
			nextID:  nextID,
		}
	}
	return store
}

func (s *Store) List(resource string) ([]map[string]any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	found, ok := s.resources[resource]
	if !ok {
		return nil, false
	}

	return copyRecords(found.Records), true
}

func (s *Store) Get(resource string, id string) (map[string]any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	found, ok := s.resources[resource]
	if !ok {
		return nil, false
	}

	index := findRecordIndex(found.Records, id)
	if index == -1 {
		return nil, false
	}

	return copyRecord(found.Records[index]), true
}

func (s *Store) Create(resource string, record map[string]any) (map[string]any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	found, ok := s.resources[resource]
	if !ok {
		return nil, false
	}

	next := copyRecord(record)

	if _, ok := next["id"]; !ok {
		next["id"] = found.nextID
		found.nextID++
	} else if id, ok := intID(next["id"]); ok && id >= found.nextID {
		found.nextID = id + 1
	}

	found.Records = append(found.Records, next)
	return copyRecord(next), true
}

func (s *Store) Replace(resource string, id string, record map[string]any) (map[string]any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	found, ok := s.resources[resource]
	if !ok {
		return nil, false
	}

	index := findRecordIndex(found.Records, id)
	if index == -1 {
		return nil, false
	}

	next := copyRecord(record)
	next["id"] = found.Records[index]["id"]

	found.Records[index] = next
	return copyRecord(next), true
}

func (s *Store) Patch(resource string, id string, patch map[string]any) (map[string]any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	found, ok := s.resources[resource]
	if !ok {
		return nil, false
	}

	index := findRecordIndex(found.Records, id)
	if index == -1 {
		return nil, false
	}

	next := copyRecord(found.Records[index])
	for key, value := range patch {
		if key == "id" {
			continue
		}
		next[key] = value
	}

	found.Records[index] = next
	return copyRecord(next), true
}

func (s *Store) Delete(resource string, id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	found, ok := s.resources[resource]
	if !ok {
		return false
	}

	index := findRecordIndex(found.Records, id)
	if index == -1 {
		return false
	}

	found.Records = append(found.Records[:index], found.Records[index+1:]...)
	return true
}

func findRecordIndex(records []map[string]any, id string) int {
	for index, record := range records {
		if fmt.Sprint(record["id"]) == id {
			return index
		}
	}
	return -1
}

func copyRecords(records []map[string]any) []map[string]any {
	copied := make([]map[string]any, 0, len(records))

	for _, record := range records {
		copied = append(copied, copyRecord(record))
	}

	return copied
}

func copyRecord(record map[string]any) map[string]any {
	copied := make(map[string]any, len(record))

	for key, value := range record {
		copied[key] = value
	}

	return copied
}

func intID(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case float64:
		return int(typed), typed == float64(int(typed))
	case string:
		id, err := strconv.Atoi(typed)
		return id, err == nil
	default:
		return 0, false
	}
}
