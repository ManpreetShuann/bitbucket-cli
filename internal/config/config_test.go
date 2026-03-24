package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_EnvVarsOverride(t *testing.T) {
	t.Setenv("BITBUCKET_URL", "https://env.example.com")
	t.Setenv("BITBUCKET_TOKEN", "env-token")

	cfg, err := Load("default", "")
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.URL != "https://env.example.com" {
		t.Errorf("expected env URL, got %s", cfg.URL)
	}
	if cfg.Token != "env-token" {
		t.Errorf("expected env token, got %s", cfg.Token)
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	credsPath := filepath.Join(dir, "credentials.yaml")

	os.WriteFile(configPath, []byte(`
current-profile: default
profiles:
  default:
    url: https://file.example.com
    default-project: PROJ
    default-repo: my-repo
`), 0644)

	os.WriteFile(credsPath, []byte(`
profiles:
  default:
    token: file-token
`), 0600)

	cfg, err := LoadFromDir(dir, "default")
	if err != nil {
		t.Fatalf("LoadFromDir error: %v", err)
	}
	if cfg.URL != "https://file.example.com" {
		t.Errorf("expected file URL, got %s", cfg.URL)
	}
	if cfg.Token != "file-token" {
		t.Errorf("expected file token, got %s", cfg.Token)
	}
	if cfg.DefaultProject != "PROJ" {
		t.Errorf("expected default project PROJ, got %s", cfg.DefaultProject)
	}
}
