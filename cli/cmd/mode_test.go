package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestModeCommandWritesSessionModeAndPreservesGuard(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("config dir is APPDATA on windows; XDG override not honored")
	}
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("HOME", tmp)

	cfgDir := filepath.Join(tmp, "ifly")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	initial := "version: 1\nsession_guard_override: open\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "state.yaml"), []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"mode", "silent"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "IFLy mode -> silent") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	data, err := os.ReadFile(filepath.Join(cfgDir, "state.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	if !strings.Contains(body, "active_mode: silent") || !strings.Contains(body, "session_guard_override: open") {
		t.Fatalf("state was not preserved correctly: %s", body)
	}
}

func TestModeCommandProjectWritesProjectState(t *testing.T) {
	tmp := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"mode", "--project", "verbose"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "IFLy mode -> verbose (project)") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	data, err := os.ReadFile(filepath.Join(tmp, ".ifly-state.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "active_mode: verbose") {
		t.Fatalf("project state missing mode: %s", data)
	}
}
