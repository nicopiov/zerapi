package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ServeConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Readonly bool   `json:"readonly" yaml:"readonly"`
	Watch    bool   `json:"watch" yaml:"watch"`
	CORS     bool   `json:"cors" yaml:"cors"`
	Delay    string `json:"delay" yaml:"delay"`
}

func Load(path string) (*ServeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var config ServeConfig
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("parse json config: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("parse yaml config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file type: %s", filepath.Ext(path))
	}
	return &config, nil
}
