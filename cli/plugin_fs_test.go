package main

import (
	"io/fs"
	"testing"
)

func TestPluginFSContainsManifest(t *testing.T) {
	data, err := fs.ReadFile(PluginFS, "plugin/.claude-plugin/plugin.json")
	if err != nil {
		t.Fatalf("read plugin.json: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("plugin.json is empty")
	}
}

func TestPluginFSContainsDefaults(t *testing.T) {
	if _, err := fs.ReadFile(PluginFS, "plugin/defaults.yaml"); err != nil {
		t.Fatalf("read defaults.yaml: %v", err)
	}
}

func TestPluginFSContainsGuardHook(t *testing.T) {
	// Hook scripts use extensionless names so Claude Code's Windows
	// auto-prepend-bash detection doesn't munge the command line.
	if _, err := fs.ReadFile(PluginFS, "plugin/hooks/guard"); err != nil {
		t.Fatalf("read guard: %v", err)
	}
}

func TestPluginFSContainsHooksJSON(t *testing.T) {
	if _, err := fs.ReadFile(PluginFS, "plugin/hooks/hooks.json"); err != nil {
		t.Fatalf("read hooks.json: %v", err)
	}
}

func TestPluginFSContainsAllFiveSkills(t *testing.T) {
	modes := []string{"silent", "minimal", "normal", "verbose", "caveman"}
	for _, m := range modes {
		p := "plugin/skills/ifly-mode-" + m + "/SKILL.md"
		if _, err := fs.ReadFile(PluginFS, p); err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
	}
}
