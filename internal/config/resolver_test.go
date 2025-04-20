package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveConfigPath(t *testing.T) {
	t.Setenv("YXA_CONFIG", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	home, _ := os.UserHomeDir()
	_ = os.Remove(filepath.Join(home, ".yxa.yml"))

	// 1. Flag
	flag := "/tmp/test-flag.yml"
	if err := os.WriteFile(flag, []byte("test: flag"), 0644); err != nil {
		t.Fatalf("Failed to write flag file: %v", err)
	}
	path, err := ResolveConfigPath(flag)
	if err != nil || path != flag {
		t.Errorf("flag: got %v, want %v, err=%v", path, flag, err)
	}
	_ = os.Remove(flag)

	// 2. Env var
	if err := os.WriteFile("/tmp/test-env.yml", []byte("test: env"), 0644); err != nil {
		t.Fatalf("Failed to write /tmp/test-env.yml: %v", err)
	}
	t.Setenv("YXA_CONFIG", "/tmp/test-env.yml")
	path, err = ResolveConfigPath("")
	if err != nil || path != "/tmp/test-env.yml" {
		t.Errorf("env: got %v, want %v, err=%v", path, "/tmp/test-env.yml", err)
	}
	_ = os.Remove("/tmp/test-env.yml")
	t.Setenv("YXA_CONFIG", "")

	// 3. Local
	if err := os.WriteFile("yxa.yml", []byte("test: local"), 0644); err != nil {
		t.Fatalf("Failed to write yxa.yml: %v", err)
	}
	path, err = ResolveConfigPath("")
	if err != nil || path != "yxa.yml" {
		t.Errorf("local: got %v, want yxa.yml, err=%v", path, err)
	}
	_ = os.Remove("yxa.yml")

	// 4. XDG
	t.Setenv("XDG_CONFIG_HOME", "/tmp")
	if err := os.MkdirAll("/tmp/yxa", 0755); err != nil {
		t.Fatalf("Failed to mkdir /tmp/yxa: %v", err)
	}
	if err := os.WriteFile("/tmp/yxa/config.yml", []byte("test: xdg"), 0644); err != nil {
		t.Fatalf("Failed to write /tmp/yxa/config.yml: %v", err)
	}
	path, err = ResolveConfigPath("")
	if err != nil || path != "/tmp/yxa/config.yml" {
		t.Errorf("xdg: got %v, want /tmp/yxa/config.yml, err=%v", path, err)
	}
	_ = os.Remove("/tmp/yxa/config.yml")
	_ = os.RemoveAll("/tmp/yxa")
	t.Setenv("XDG_CONFIG_HOME", "")

	// 5. Home
	if err := os.WriteFile(filepath.Join(home, ".yxa.yml"), []byte("test: home"), 0644); err != nil {
		t.Fatalf("Failed to write home yxa.yml: %v", err)
	}
	path, err = ResolveConfigPath("")
	if err != nil || path != filepath.Join(home, ".yxa.yml") {
		t.Errorf("home: got %v, want %v, err=%v", path, filepath.Join(home, ".yxa.yml"), err)
	}
	_ = os.Remove(filepath.Join(home, ".yxa.yml"))

	// 6. Not found
	path, err = ResolveConfigPath("")
	if err == nil {
		t.Errorf("not found: expected error, got %v", path)
	}
}
