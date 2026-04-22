package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestDefaultsDrift ensures the Go unmarshal of defaults.yaml matches the
// values reported by hooks/lib/parse_defaults.sh for a fixed set of keys.
//
// parse_defaults.sh is a stdin→stdout filter that emits "dotted.key=value"
// lines. We invoke it once, scrape the expected keys, and compare against
// what the Go unmarshaler produced.
func TestDefaultsDrift(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash reader not exercised on windows CI")
	}
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Skip("repo root not found; drift check only runs from working tree:", err)
	}
	defaultsPath := filepath.Join(repoRoot, "defaults.yaml")
	data, err := os.ReadFile(defaultsPath)
	if err != nil {
		t.Fatalf("read defaults.yaml: %v", err)
	}
	goCfg, err := LoadBytes(data)
	if err != nil {
		t.Fatalf("go parse: %v", err)
	}

	readerPath := filepath.Join(repoRoot, "hooks", "lib", "parse_defaults.sh")
	if _, err := os.Stat(readerPath); err != nil {
		t.Skip("parse_defaults.sh not present")
	}

	cmd := exec.Command("bash", readerPath)
	cmd.Stdin = strings.NewReader(string(data))
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("bash reader failed: %v", err)
	}
	bashKV := parseKVLines(string(out))

	checks := map[string]string{
		"mode":        goCfg.Mode,
		"guard.level": goCfg.Guard.Level,
		"version":     fmt.Sprintf("%d", goCfg.Version),
	}
	for key, goVal := range checks {
		bashVal, ok := bashKV[key]
		if !ok {
			t.Errorf("bash reader did not emit %q; emitted: %v", key, bashKV)
			continue
		}
		if bashVal != goVal {
			t.Errorf("drift at %q: go=%q bash=%q", key, goVal, bashVal)
		}
	}
}

func parseKVLines(s string) map[string]string {
	m := map[string]string{}
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		// Skip list entries (key[n]=v) — only compare scalars.
		if strings.Contains(line, "[") && strings.Contains(line, "]=") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		m[line[:idx]] = line[idx+1:]
	}
	return m
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "defaults.yaml")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", os.ErrNotExist
}
