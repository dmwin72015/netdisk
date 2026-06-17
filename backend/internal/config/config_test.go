package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadResolvesStorageRootRelativeToConfigFile(t *testing.T) {
	// Required by validateSecrets — Load() now refuses to boot without them.
	t.Setenv("NETDISK_DB_DSN", "postgres://test:test@localhost:5432/test?sslmode=disable")
	t.Setenv("NETDISK_JWT_SECRET", "test-secret")

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

func TestLoadFailsWhenJWTSecretMissing(t *testing.T) {
	// Only DSN is set; JWT secret is intentionally absent.
	t.Setenv("NETDISK_DB_DSN", "postgres://test:test@localhost:5432/test?sslmode=disable")
	t.Setenv("NETDISK_JWT_SECRET", "")

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte("storage:\n  root: data\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := Load(configPath); err == nil {
		t.Fatalf("Load() expected error for missing jwt.secret, got nil")
	}
}

func TestLoadReadsOAuthProviderSecretsFromEnv(t *testing.T) {
	t.Setenv("NETDISK_DB_DSN", "postgres://test:test@localhost:5432/test?sslmode=disable")
	t.Setenv("NETDISK_JWT_SECRET", "secret")
	t.Setenv("NETDISK_OAUTH2_2LIBRA_CLIENT_ID", "id-from-env")
	t.Setenv("NETDISK_OAUTH2_2LIBRA_CLIENT_SECRET", "secret-from-env")

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	yaml := []byte(`
oauth2:
  providers:
    "2libra":
      client_id: ""
      client_secret: ""
      auth_url: "https://2libra.example/oauth/authorize"
      token_url: "https://2libra.example/oauth/token"
      user_info_url: "https://2libra.example/oauth/me"
      scope: "profile"
`)
	if err := os.WriteFile(configPath, yaml, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	prov, ok := cfg.OAuth2.Providers["2libra"]
	if !ok {
		t.Fatalf("expected 2libra provider to be loaded")
	}
	if prov.ClientID != "id-from-env" {
		t.Errorf("ClientID = %q, want %q", prov.ClientID, "id-from-env")
	}
	if prov.ClientSecret != "secret-from-env" {
		t.Errorf("ClientSecret = %q, want %q", prov.ClientSecret, "secret-from-env")
	}
	if prov.AuthURL != "https://2libra.example/oauth/authorize" {
		t.Errorf("AuthURL = %q, want yaml value preserved", prov.AuthURL)
	}
}
