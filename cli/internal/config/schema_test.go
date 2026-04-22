package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestUnmarshalFullConfig(t *testing.T) {
	raw := []byte(`
version: 1
mode: minimal
guard:
  level: strict
  lockdown: false
  tools:
    bash: true
    edit: true
    web_fetch: false
  additional_dirs:
    - /tmp/a
  blocked_commands:
    - "docker rm"
  allowed_network:
    - api.github.com
  sensitive_paths:
    - "~/.ssh/"
telemetry:
  easter_egg: true
`)
	var c Config
	if err := yaml.Unmarshal(raw, &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if c.Mode != "minimal" {
		t.Errorf("mode %q", c.Mode)
	}
	if c.Guard.Level != "strict" {
		t.Errorf("level %q", c.Guard.Level)
	}
	if c.Guard.Tools.Bash == nil || !*c.Guard.Tools.Bash {
		t.Error("bash should be true")
	}
	if c.Guard.Tools.WebFetch == nil || *c.Guard.Tools.WebFetch {
		t.Error("web_fetch should be explicitly false")
	}
	if c.Guard.Tools.Glob != nil {
		t.Error("glob was not set, should be nil")
	}
	if len(c.Guard.AdditionalDirs) != 1 || c.Guard.AdditionalDirs[0] != "/tmp/a" {
		t.Errorf("additional_dirs %v", c.Guard.AdditionalDirs)
	}
}

func TestLevelRank(t *testing.T) {
	cases := map[string]int{"strict": 3, "project": 2, "open": 1, "off": 0, "bogus": -1}
	for in, want := range cases {
		if got := LevelRank(in); got != want {
			t.Errorf("LevelRank(%q)=%d want %d", in, got, want)
		}
	}
}

func TestValidateMode(t *testing.T) {
	for _, m := range []string{"silent", "minimal", "normal", "verbose", "caveman"} {
		if err := ValidateMode(m); err != nil {
			t.Errorf("mode %q: %v", m, err)
		}
	}
	if ValidateMode("chatty") == nil {
		t.Error("expected error for invalid mode")
	}
}

func TestValidateLevel(t *testing.T) {
	for _, l := range []string{"strict", "project", "open", "off"} {
		if err := ValidateLevel(l); err != nil {
			t.Errorf("level %q: %v", l, err)
		}
	}
	if ValidateLevel("permissive") == nil {
		t.Error("expected error for invalid level")
	}
}
