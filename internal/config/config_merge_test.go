package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMergeConfigs(t *testing.T) {
	global := &ProjectConfig{
		Name: "global",
		Variables: map[string]string{
			"A": "globalA",
			"B": "globalB",
		},
		Commands: map[string]Command{
			"gcmd": {Run: "echo global"},
			"shared": {Run: "echo global-shared"},
		},
	}
	project := &ProjectConfig{
		Name: "project",
		Variables: map[string]string{
			"B": "projB",
			"C": "projC",
		},
		Commands: map[string]Command{
			"pcmd": {Run: "echo project"},
			"shared": {Run: "echo project-shared"},
		},
	}
	merged := MergeConfigs(global, project)
	if merged.Name != "project" {
		t.Errorf("Name: got %v, want project", merged.Name)
	}
	if merged.Variables["A"] != "globalA" || merged.Variables["B"] != "projB" || merged.Variables["C"] != "projC" {
		t.Errorf("Variables not merged as expected: %+v", merged.Variables)
	}
	if merged.Commands["gcmd"].Run != "echo global" || merged.Commands["pcmd"].Run != "echo project" || merged.Commands["shared"].Run != "echo project-shared" {
		t.Errorf("Commands not merged as expected: %+v", merged.Commands)
	}
}

func assertVariable(t *testing.T, got, want, name string) {
	t.Helper()
	if got != want {
		t.Errorf("Variable %s: got %q, want %q", name, got, want)
	}
}

func assertCommand(t *testing.T, got Command, wantRun, name string) {
	t.Helper()
	if got.Run != wantRun {
		t.Errorf("Command %s: got run %q, want %q", name, got.Run, wantRun)
	}
}

func TestLoadConfigFrom_MergesGlobal_ProjectOverridesGlobal(t *testing.T) {
	dir := t.TempDir()
	globalConfig := `
name: global
variables:
  A: globalA
  B: globalB
commands:
  gcmd:
    run: echo global
  shared:
    run: echo global-shared
`
	projectConfig := `
name: project
variables:
  B: projB
  C: projC
commands:
  pcmd:
    run: echo project
  shared:
    run: echo project-shared
`
	globalPath := filepath.Join(dir, ".yxa.yml")
	projectPath := filepath.Join(dir, "yxa.yml")
	if err := os.WriteFile(globalPath, []byte(globalConfig), 0644); err != nil {
		t.Fatalf("Failed to write global config: %v", err)
	}
	if err := os.WriteFile(projectPath, []byte(projectConfig), 0644); err != nil {
		t.Fatalf("Failed to write project config: %v", err)
	}
	oldHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", dir); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}
	t.Cleanup(func() { _ = os.Setenv("HOME", oldHome) })

	cfg, err := LoadConfigFrom(projectPath)
	if err != nil {
		t.Fatalf("LoadConfigFrom error: %v", err)
	}
	if cfg.Name != "project" {
		t.Errorf("Name: got %v, want project", cfg.Name)
	}
	assertVariable(t, cfg.Variables["A"], "globalA", "A")
	assertVariable(t, cfg.Variables["B"], "projB", "B")
	assertVariable(t, cfg.Variables["C"], "projC", "C")
	assertCommand(t, cfg.Commands["gcmd"], "echo global", "gcmd")
	assertCommand(t, cfg.Commands["pcmd"], "echo project", "pcmd")
	assertCommand(t, cfg.Commands["shared"], "echo project-shared", "shared")

}


