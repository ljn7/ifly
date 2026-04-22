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

func TestStatusPrintsMergedConfig(t *testing.T) {
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
	cfg := `version: 1
mode: verbose
guard:
  level: project
  lockdown: false
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}

	SetVersion("0.1.0")
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"status"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "mode: verbose") {
		t.Errorf("missing mode line: %q", out)
	}
	if !strings.Contains(out, "guard: project") {
		t.Errorf("missing guard line: %q", out)
	}
}

func TestStatusShowsLockdownViolations(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("config dir is APPDATA on windows; XDG override not honored")
	}
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("HOME", tmp)
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)
	t.Setenv("IFLY_MODE", "")
	t.Setenv("IFLY_GUARD", "off")

	cfgDir, err := paths.GlobalConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	global := `version: 1
guard:
  level: strict
  lockdown: true
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(global), 0o644); err != nil {
		t.Fatal(err)
	}

	SetVersion("0.1.0")
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"status"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "lockdown") {
		t.Errorf("expected lockdown notice, got %q", buf.String())
	}
}
