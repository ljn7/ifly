package detect

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestClaudeDirLinuxXDG(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	tmp := t.TempDir()
	xdg := filepath.Join(tmp, "xdg")
	claudeDir := filepath.Join(xdg, "claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", xdg)
	t.Setenv("HOME", tmp)

	got, err := ClaudeConfigDir(HostInfo{OS: "linux"})
	if err != nil {
		t.Fatal(err)
	}
	if got != claudeDir {
		t.Errorf("got %s want %s", got, claudeDir)
	}
}

func TestClaudeDirLinuxFallbackDotClaude(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	tmp := t.TempDir()
	dotClaude := filepath.Join(tmp, ".claude")
	if err := os.MkdirAll(dotClaude, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", tmp)

	got, err := ClaudeConfigDir(HostInfo{OS: "linux"})
	if err != nil {
		t.Fatal(err)
	}
	if got != dotClaude {
		t.Errorf("got %s want %s", got, dotClaude)
	}
}

func TestClaudeDirNotFound(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("HOME", tmp)

	_, err := ClaudeConfigDir(HostInfo{OS: "linux"})
	if err == nil {
		t.Fatal("expected error when no candidate exists")
	}
}

func TestClaudeDirDarwin(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	want := filepath.Join(tmp, "Library", "Application Support", "claude")
	if err := os.MkdirAll(want, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := ClaudeConfigDir(HostInfo{OS: "darwin"})
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got %s want %s", got, want)
	}
}
