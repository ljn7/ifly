package paths

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestGlobalConfigDirXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/xdg/cfg")
	t.Setenv("HOME", "/home/u")
	if runtime.GOOS == "windows" {
		t.Skip("XDG not used on windows")
	}
	got, err := GlobalConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/xdg/cfg", "ifly")
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestGlobalConfigDirHomeFallback(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("XDG not used on windows")
	}
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/home/u")
	got, err := GlobalConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/home/u", ".config", "ifly")
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestStateFileUnderConfigDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("path format differs on windows")
	}
	t.Setenv("XDG_CONFIG_HOME", "/xdg/cfg")
	got, err := StateFile()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/xdg/cfg", "ifly", "state.yaml")
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestGlobalConfigFileName(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	t.Setenv("XDG_CONFIG_HOME", "/x")
	got, err := GlobalConfigFile()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(got) != "config.yaml" {
		t.Errorf("expected config.yaml, got %s", got)
	}
}
