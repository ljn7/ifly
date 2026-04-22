package installer

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func fakePluginFS() fs.FS {
	return fstest.MapFS{
		"plugin/.claude-plugin/plugin.json":       &fstest.MapFile{Data: []byte(`{"name":"ifly"}`)},
		"plugin/defaults.yaml":                    &fstest.MapFile{Data: []byte("version: 1\n")},
		"plugin/hooks/guard.sh":                   &fstest.MapFile{Data: []byte("#!/usr/bin/env bash\n"), Mode: 0o755},
		"plugin/hooks/ifly-state":                 &fstest.MapFile{Data: []byte("#!/usr/bin/env bash\n"), Mode: 0o755},
		"plugin/hooks/lib/path_resolve.sh":        &fstest.MapFile{Data: []byte("# lib\n")},
		"plugin/skills/ifly-mode-silent/SKILL.md": &fstest.MapFile{Data: []byte("# silent\n")},
		"plugin/commands/guard.md":                &fstest.MapFile{Data: []byte("# guard\n")},
		"plugin/commands/mode.md":                 &fstest.MapFile{Data: []byte("# mode\n")},
		"plugin/commands/status.md":               &fstest.MapFile{Data: []byte("# status\n")},
	}
}

func TestInstallCopiesAllPluginFiles(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "claude", "plugins", "ifly")
	opts := Options{Dest: dst, Overwrite: false}
	if err := Install(fakePluginFS(), opts); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{
		".claude-plugin/plugin.json",
		"defaults.yaml",
		"hooks/guard.sh",
		"hooks/ifly-state",
		"hooks/lib/path_resolve.sh",
		"skills/ifly-mode-silent/SKILL.md",
		"commands/guard.md",
		"commands/mode.md",
		"commands/status.md",
	} {
		if _, err := os.Stat(filepath.Join(dst, rel)); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}
}

func TestInstallPreservesShellExecutableBit(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("executable bit not tracked on windows")
	}
	dst := filepath.Join(t.TempDir(), "claude")
	if err := Install(fakePluginFS(), Options{Dest: dst, Overwrite: false}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(dst, "hooks", "guard.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("guard.sh not executable: %v", info.Mode())
	}
}

func TestInstallRefusesWithoutOverwrite(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "claude")
	if err := Install(fakePluginFS(), Options{Dest: dst}); err != nil {
		t.Fatal(err)
	}
	if err := Install(fakePluginFS(), Options{Dest: dst, Overwrite: false}); err == nil {
		t.Fatal("expected error on second install without overwrite")
	}
}

func TestInstallOverwritesWithFlag(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "claude")
	if err := Install(fakePluginFS(), Options{Dest: dst}); err != nil {
		t.Fatal(err)
	}
	if err := Install(fakePluginFS(), Options{Dest: dst, Overwrite: true}); err != nil {
		t.Fatal(err)
	}
}

func TestExtensionlessHookGetsExecutableBit(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("executable bit not tracked on windows")
	}
	tree := fstest.MapFS{
		"plugin/hooks/guard":         &fstest.MapFile{Data: []byte("#!/usr/bin/env bash\n")},
		"plugin/hooks/session-start": &fstest.MapFile{Data: []byte("#!/usr/bin/env bash\n")},
		"plugin/hooks/hooks.json":    &fstest.MapFile{Data: []byte("{}\n")},
		"plugin/hooks/lib/x.sh":      &fstest.MapFile{Data: []byte("# lib\n")},
	}
	dst := filepath.Join(t.TempDir(), "out")
	if err := Install(tree, Options{Dest: dst}); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"hooks/guard", "hooks/session-start"} {
		info, err := os.Stat(filepath.Join(dst, rel))
		if err != nil {
			t.Fatalf("stat %s: %v", rel, err)
		}
		if info.Mode().Perm()&0o100 == 0 {
			t.Errorf("%s should be executable, mode=%v", rel, info.Mode())
		}
	}
	info, err := os.Stat(filepath.Join(dst, "hooks", "hooks.json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o100 != 0 {
		t.Errorf("hooks.json should NOT be executable, mode=%v", info.Mode())
	}
}
