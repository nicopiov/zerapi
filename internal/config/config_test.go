package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadJSONConfig(t *testing.T) {
	path := writeTempFile(t, "zerapi.json", `{
		"host": "0.0.0.0",
		"port": 9090,
		"readonly": true,
		"watch": true,
		"cors": true,
		"delay": "250ms"
	}`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Host != "0.0.0.0" {
		t.Fatalf("expected host 0.0.0.0, got %q", cfg.Host)
	}

	if cfg.Port != 9090 {
		t.Fatalf("expected port 9090, got %d", cfg.Port)
	}

	if !cfg.Readonly || !cfg.Watch || !cfg.CORS {
		t.Fatalf("expected boolean config values to be true")
	}

	if cfg.Delay != "250ms" {
		t.Fatalf("expected delay 250ms, got %q", cfg.Delay)
	}
}

func TestLoadYAMLConfig(t *testing.T) {
	path := writeTempFile(t, "zerapi.yaml", `
host: 0.0.0.0
port: 9090
readonly: true
watch: true
cors: true
delay: 250ms
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Host != "0.0.0.0" {
		t.Fatalf("expected host 0.0.0.0, got %q", cfg.Host)
	}

	if cfg.Port != 9090 {
		t.Fatalf("expected port 9090, got %d", cfg.Port)
	}

	if !cfg.Readonly || !cfg.Watch || !cfg.CORS {
		t.Fatalf("expected boolean config values to be true")
	}

	if cfg.Delay != "250ms" {
		t.Fatalf("expected delay 250ms, got %q", cfg.Delay)
	}
}

func TestLoadRejectsUnsupportedConfigType(t *testing.T) {
	path := writeTempFile(t, "zerapi.txt", ``)

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
