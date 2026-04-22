package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ljn7/ifly/cli/internal/paths"
)

func TestGuardCommandWritesSessionOverride(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("config dir is APPDATA on windows; XDG override not honored")
	}
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("HOME", tmp)
	t.Setenv("IFLY_MODE", "")
	t.Setenv("IFLY_GUARD", "")

	cfgDir, err := paths.GlobalConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("version: 1\nguard:\n  level: strict\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"guard", "open"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "IFLy guard -> open") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	statePath, err := paths.StateFile()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "session_guard_override: open") {
		t.Fatalf("missing override in state: %s", data)
	}
}

func TestGuardCommandRespectsLockdown(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("config dir is APPDATA on windows; XDG override not honored")
	}
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("HOME", tmp)
	t.Setenv("IFLY_MODE", "")
	t.Setenv("IFLY_GUARD", "")

	cfgDir, err := paths.GlobalConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := "version: 1\nguard:\n  level: strict\n  lockdown: true\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"guard", "off"})
	err = rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "lockdown enabled") {
		t.Fatalf("expected lockdown error, got %v", err)
	}
}

func TestGuardCommandProjectWritesProjectState(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("global config dir is APPDATA on windows; XDG override not honored")
	}
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "xdg"))
	t.Setenv("HOME", tmp)
	t.Setenv("IFLY_MODE", "")
	t.Setenv("IFLY_GUARD", "")
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)

	cfgDir, err := paths.GlobalConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("version: 1\nguard:\n  level: strict\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"guard", "--project", "project"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "IFLy guard -> project (project)") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	data, err := os.ReadFile(filepath.Join(tmp, ".ifly-state.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "session_guard_override: project") {
		t.Fatalf("project state missing guard: %s", data)
	}
}
