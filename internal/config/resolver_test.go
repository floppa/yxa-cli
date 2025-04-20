package config

import (
	"os"
	"path/filepath"
	"testing"
)

// Helper to isolate environment and HOME for each test
func withIsolatedEnv(t *testing.T, testFunc func(home string)) {
	t.Setenv("YXA_CONFIG", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	testFunc(home)
}

func TestResolveConfigPath_Flag(t *testing.T) {
	withIsolatedEnv(t, func(_ string) {
		path := "/tmp/test-flag.yml"
		if err := os.WriteFile(path, []byte("testflag"), 0644); err != nil {
			t.Fatalf("failed to write flag config%v", err)
		}
		defer func() {
			if err := os.Remove(path); err != nil {
				t.Errorf("failed to remove flag config%v", err)
			}
		}()
		got, err := ResolveConfigPath(path)
		if err != nil {
			t.Errorf("unexpected error%v", err)
		}
		if got != path {
			t.Errorf("got %v, want %v", got, path)
		}
	})
}

func TestResolveConfigPath_Env(t *testing.T) {
	withIsolatedEnv(t, func(_ string) {
		path := "/tmp/test-env.yml"
		if err := os.WriteFile(path, []byte("testenv"), 0644); err != nil {
			t.Fatalf("failed to write env config%v", err)
		}
		t.Setenv("YXA_CONFIG", path)
		defer func() {
			if err := os.Remove(path); err != nil {
				t.Errorf("failed to remove env config%v", err)
			}
		}()
		got, err := ResolveConfigPath("")
		if err != nil {
			t.Errorf("unexpected error%v", err)
		}
		if got != path {
			t.Errorf("got %v, want %v", got, path)
		}
	})
}

func TestResolveConfigPath_Local(t *testing.T) {
	withIsolatedEnv(t, func(_ string) {
		path := "yxa.yml"
		if err := os.WriteFile(path, []byte("testlocal"), 0644); err != nil {
			t.Fatalf("failed to write local config%v", err)
		}
		defer func() {
			if err := os.Remove(path); err != nil {
				t.Errorf("failed to remove local config%v", err)
			}
		}()
		got, err := ResolveConfigPath("")
		if err != nil {
			t.Errorf("unexpected error%v", err)
		}
		if got != path {
			t.Errorf("got %v, want %v", got, path)
		}
	})
}

func TestResolveConfigPath_XDG(t *testing.T) {
	withIsolatedEnv(t, func(_ string) {
		xdg := "/tmp"
		t.Setenv("XDG_CONFIG_HOME", xdg)
		dir := filepath.Join(xdg, "yxa")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to mkdir %s%v", dir, err)
		}
		cfgPath := filepath.Join(dir, "config.yml")
		if err := os.WriteFile(cfgPath, []byte("testxdg"), 0644); err != nil {
			t.Fatalf("failed to write xdg config%v", err)
		}
		defer func() {
			if err := os.Remove(cfgPath); err != nil {
				t.Errorf("failed to remove xdg config%v", err)
			}
			if err := os.RemoveAll(dir); err != nil {
				t.Errorf("failed to remove %s%v", dir, err)
			}
		}()
		got, err := ResolveConfigPath("")
		if err != nil {
			t.Errorf("unexpected error%v", err)
		}
		if got != cfgPath {
			t.Errorf("got %v, want %v", got, cfgPath)
		}
	})
}

func TestResolveConfigPath_Home(t *testing.T) {
	withIsolatedEnv(t, func(home string) {
		cfgPath := filepath.Join(home, ".yxa.yml")
		if err := os.WriteFile(cfgPath, []byte("testhome"), 0644); err != nil {
			t.Fatalf("failed to write home config%v", err)
		}
		defer func() {
			if err := os.Remove(cfgPath); err != nil {
				t.Errorf("failed to remove home config%v", err)
			}
		}()
		got, err := ResolveConfigPath("")
		if err != nil {
			t.Errorf("unexpected error%v", err)
		}
		if got != cfgPath {
			t.Errorf("got %v, want %v", got, cfgPath)
		}
	})
}

func TestResolveConfigPath_NotFound(t *testing.T) {
	withIsolatedEnv(t, func(_ string) {
		_, err := ResolveConfigPath("")
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}
