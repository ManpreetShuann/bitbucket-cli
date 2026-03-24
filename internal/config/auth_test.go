package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadCredentials(t *testing.T) {
	dir := t.TempDir()
	credsPath := filepath.Join(dir, "credentials.yaml")

	err := SaveCredentials(credsPath, "test-profile", "my-token")
	if err != nil {
		t.Fatalf("SaveCredentials error: %v", err)
	}

	info, err := os.Stat(credsPath)
	if err != nil {
		t.Fatalf("Stat error: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600 permissions, got %o", info.Mode().Perm())
	}

	token, err := LoadToken(credsPath, "test-profile")
	if err != nil {
		t.Fatalf("LoadToken error: %v", err)
	}
	if token != "my-token" {
		t.Errorf("expected my-token, got %s", token)
	}
}

func TestRemoveCredentials(t *testing.T) {
	dir := t.TempDir()
	credsPath := filepath.Join(dir, "credentials.yaml")

	if err := SaveCredentials(credsPath, "test-profile", "my-token"); err != nil {
		t.Fatalf("SaveCredentials error: %v", err)
	}
	err := RemoveCredentials(credsPath, "test-profile")
	if err != nil {
		t.Fatalf("RemoveCredentials error: %v", err)
	}

	token, err := LoadToken(credsPath, "test-profile")
	if err != nil {
		t.Fatalf("LoadToken error: %v", err)
	}
	if token != "" {
		t.Errorf("expected empty token after remove, got %s", token)
	}
}
