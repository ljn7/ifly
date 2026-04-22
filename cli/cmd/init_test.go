package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitWritesIflyYAML(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"init"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(tmp, ".ifly.yaml"))
	if err != nil {
		t.Fatalf("expected .ifly.yaml: %v", err)
	}
	if !strings.Contains(string(data), "version: 1") {
		t.Errorf("template missing version: %s", data)
	}
	if !strings.Contains(string(data), "guard:") {
		t.Errorf("template missing guard section: %s", data)
	}
}

func TestInitDoesNotOverwriteWithoutForce(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)
	p := filepath.Join(tmp, ".ifly.yaml")
	_ = os.WriteFile(p, []byte("# hand-edited\n"), 0o644)

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"init"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected refusal without --force")
	}
	data, _ := os.ReadFile(p)
	if !strings.Contains(string(data), "hand-edited") {
		t.Errorf("existing file was overwritten: %s", data)
	}
}

func TestInitForceOverwrites(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)
	p := filepath.Join(tmp, ".ifly.yaml")
	_ = os.WriteFile(p, []byte("# hand-edited\n"), 0o644)

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"init", "--force"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(p)
	if strings.Contains(string(data), "hand-edited") {
		t.Error("expected overwrite with --force")
	}
}
