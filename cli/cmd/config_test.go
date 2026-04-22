package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func setupConfigEnv(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("config dir is APPDATA on windows; XDG override not honored")
	}
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("HOME", tmp)
	t.Setenv("IFLY_MODE", "")
	t.Setenv("IFLY_GUARD", "")
	return tmp
}

func TestConfigShowPrintsMerged(t *testing.T) {
	tmp := setupConfigEnv(t)
	cfgDir := filepath.Join(tmp, "ifly")
	_ = os.MkdirAll(cfgDir, 0o755)
	_ = os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("version: 1\nmode: minimal\n"), 0o644)

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"config", "show"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "mode: minimal") {
		t.Errorf("show output missing mode: %q", buf.String())
	}
}

func TestConfigSetWritesGlobal(t *testing.T) {
	tmp := setupConfigEnv(t)
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "set", "mode", "verbose"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(tmp, "ifly", "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "mode: verbose") {
		t.Errorf("config.yaml did not persist: %s", data)
	}
}

func TestConfigSetRejectsInvalidMode(t *testing.T) {
	setupConfigEnv(t)
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "set", "mode", "chatty"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestConfigGet(t *testing.T) {
	tmp := setupConfigEnv(t)
	cfgDir := filepath.Join(tmp, "ifly")
	_ = os.MkdirAll(cfgDir, 0o755)
	_ = os.WriteFile(filepath.Join(cfgDir, "config.yaml"),
		[]byte("version: 1\nguard:\n  level: project\n"), 0o644)

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"config", "get", "guard.level"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buf.String()) != "project" {
		t.Errorf("got %q", buf.String())
	}
}
