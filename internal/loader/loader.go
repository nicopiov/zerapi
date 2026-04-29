package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Resource struct {
	Name    string
	Records []map[string]any
}

type Result struct {
	Resources []Resource
}

func Load(path string) (*Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	extension := strings.ToLower(filepath.Ext(path))
	var value any
	switch extension {
	case ".json":
		if err := json.Unmarshal(data, &value); err != nil {
			return nil, fmt.Errorf("parse json: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &value); err != nil {
			return nil, fmt.Errorf("parse yaml: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file type: %s", extension)
	}

	switch typed := value.(type) {
	case []any:
		records, err := recordsFromArray(typed)
		if err != nil {
			return nil, err
		}

		return &Result{
			Resources: []Resource{
				{
					Name:    resourceNameFromPath(path),
					Records: records,
				},
			},
		}, nil
	case map[string]any:
		resources := make([]Resource, 0, len(typed))

		for name, raw := range typed {
			items, ok := raw.([]any)
			if !ok {
				return nil, fmt.Errorf("resource %q must be an array", name)
			}

			records, err := recordsFromArray(items)
			if err != nil {
				return nil, fmt.Errorf("resource %q: %w", name, err)
			}

			resources = append(resources, Resource{
				Name:    name,
				Records: records,
			})
		}
		return &Result{
			Resources: resources,
		}, nil
	default:
		return nil, fmt.Errorf("file root must be an array or object")
	}
}

func recordsFromArray(items []any) ([]map[string]any, error) {
	records := make([]map[string]any, 0, len(items))

	for index, item := range items {
		record, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("record at index %d must be an object", index)
		}
		records = append(records, record)
	}
	return records, nil
}

func resourceNameFromPath(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}
