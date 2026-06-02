package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadResolvesStorageRootRelativeToConfigFile(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "backend")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	configPath := filepath.Join(configDir, "config.yaml")
	contents := []byte(`
storage:
  root: "../data"
`)
	if err := os.WriteFile(configPath, contents, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := filepath.Join(dir, "data")
	if cfg.Storage.Root != want {
		t.Fatalf("Storage.Root = %q, want %q", cfg.Storage.Root, want)
	}
}
