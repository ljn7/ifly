package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileMissingReturnsZero(t *testing.T) {
	c, err := LoadFile(filepath.Join(t.TempDir(), "nope.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if c.Mode != "" || c.Guard.Level != "" {
		t.Errorf("expected zero value, got %+v", c)
	}
}

func TestLoadFileValid(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "c.yaml")
	if err := os.WriteFile(p, []byte("version: 1\nmode: verbose\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if c.Mode != "verbose" {
		t.Errorf("mode %q", c.Mode)
	}
}

func TestLoadFileInvalidYAML(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "c.yaml")
	if err := os.WriteFile(p, []byte(":\n::: bad"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadFile(p); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("IFLY_MODE", "silent")
	t.Setenv("IFLY_GUARD", "open")
	c := LoadEnv()
	if c.Mode != "silent" {
		t.Errorf("mode %q", c.Mode)
	}
	if c.Guard.Level != "open" {
		t.Errorf("level %q", c.Guard.Level)
	}
}

func TestLoadEnvUnset(t *testing.T) {
	t.Setenv("IFLY_MODE", "")
	t.Setenv("IFLY_GUARD", "")
	c := LoadEnv()
	if c.Mode != "" || c.Guard.Level != "" {
		t.Errorf("expected empty, got %+v", c)
	}
}
