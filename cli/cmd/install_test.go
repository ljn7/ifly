package cmd

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"

	"github.com/ljn7/ifly/cli/internal/paths"
)

func fakePluginFS() fs.FS {
	return fstest.MapFS{
		"plugin/.claude-plugin/plugin.json": &fstest.MapFile{Data: []byte(`{"name":"ifly"}`)},
		"plugin/defaults.yaml":              &fstest.MapFile{Data: []byte("version: 1\n")},
		"plugin/hooks/guard.sh":             &fstest.MapFile{Data: []byte("#!/usr/bin/env bash\nexit 0\n"), Mode: 0o755},
		"plugin/hooks/ifly-state":           &fstest.MapFile{Data: []byte("#!/usr/bin/env bash\n"), Mode: 0o755},
		"plugin/commands/guard.md":          &fstest.MapFile{Data: []byte("# guard\n")},
		"plugin/commands/mode.md":           &fstest.MapFile{Data: []byte("# mode\n")},
		"plugin/commands/status.md":         &fstest.MapFile{Data: []byte("# status\n")},
	}
}

func TestInstallHeadlessWritesConfigAndPlugin(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("global config dir is APPDATA on windows; XDG override not honored")
	}
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("HOME", tmp)

	claudeDir := filepath.Join(tmp, "claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	opts := InstallOpts{
		PluginFS:  fakePluginFS(),
		ClaudeDir: claudeDir,
		Mode:      "minimal",
		Guard:     "strict",
		Lockdown:  false,
		Overwrite: false,
	}
	buf := &bytes.Buffer{}
	if err := runInstall(buf, opts); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(claudeDir, "plugins", "ifly", ".claude-plugin", "plugin.json")); err != nil {
		t.Errorf("plugin.json not written: %v", err)
	}
	globalPath, err := paths.GlobalConfigFile()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(globalPath); err != nil {
		t.Errorf("global config.yaml not written: %v", err)
	}
}

func TestInstallHeadlessRejectsInvalidMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("global config dir is APPDATA on windows; XDG override not honored")
	}
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("HOME", tmp)

	opts := InstallOpts{
		PluginFS:  fakePluginFS(),
		ClaudeDir: filepath.Join(tmp, "claude"),
		Mode:      "chatty",
		Guard:     "strict",
	}
	buf := &bytes.Buffer{}
	if err := runInstall(buf, opts); err == nil {
		t.Fatal("expected validation error")
	}
}
