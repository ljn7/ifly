package config

import (
	"reflect"
	"sort"
	"testing"
)

func b(v bool) *bool { return &v }

func TestMergeScalarReplace(t *testing.T) {
	base := Config{Mode: "minimal", Guard: Guard{Level: "strict"}}
	over := Config{Mode: "verbose"}
	got := Merge(base, over)
	if got.Mode != "verbose" {
		t.Errorf("mode %q", got.Mode)
	}
	if got.Guard.Level != "strict" {
		t.Errorf("level %q", got.Guard.Level)
	}
}

func TestMergeToolsPerKey(t *testing.T) {
	base := Config{Guard: Guard{Tools: Tools{Bash: b(true), Edit: b(true), Glob: b(false)}}}
	over := Config{Guard: Guard{Tools: Tools{Bash: b(false)}}}
	got := Merge(base, over)
	if *got.Guard.Tools.Bash != false {
		t.Error("bash should be overridden to false")
	}
	if *got.Guard.Tools.Edit != true {
		t.Error("edit should stay true from base")
	}
	if *got.Guard.Tools.Glob != false {
		t.Error("glob should stay false from base")
	}
}

func TestMergeListsAdditiveUnique(t *testing.T) {
	base := Config{Guard: Guard{AdditionalDirs: []string{"/tmp/a", "/tmp/b"}}}
	over := Config{Guard: Guard{AdditionalDirs: []string{"/tmp/b", "/tmp/c"}}}
	got := Merge(base, over)
	sort.Strings(got.Guard.AdditionalDirs)
	want := []string{"/tmp/a", "/tmp/b", "/tmp/c"}
	if !reflect.DeepEqual(got.Guard.AdditionalDirs, want) {
		t.Errorf("got %v want %v", got.Guard.AdditionalDirs, want)
	}
}

func TestMergeLockdownBlocksLoosening(t *testing.T) {
	global := Config{Guard: Guard{Level: "strict", Lockdown: true}}
	project := Config{Guard: Guard{Level: "open"}}
	got, violations := MergeRespectingLockdown(global, project, Config{})
	if got.Guard.Level != "strict" {
		t.Errorf("lockdown should keep level=strict, got %q", got.Guard.Level)
	}
	if len(violations) != 1 {
		t.Errorf("expected 1 violation, got %v", violations)
	}
}

func TestMergeLockdownAllowsTightening(t *testing.T) {
	global := Config{Guard: Guard{Level: "project", Lockdown: true}}
	project := Config{Guard: Guard{Level: "strict"}}
	got, violations := MergeRespectingLockdown(global, project, Config{})
	if got.Guard.Level != "strict" {
		t.Errorf("tightening allowed; got %q", got.Guard.Level)
	}
	if len(violations) != 0 {
		t.Errorf("no violations expected, got %v", violations)
	}
}

func TestMergeEnvHasHighestPriority(t *testing.T) {
	global := Config{Mode: "minimal", Guard: Guard{Level: "strict"}}
	project := Config{Mode: "verbose"}
	env := Config{Mode: "silent", Guard: Guard{Level: "off"}}
	got, _ := MergeRespectingLockdown(global, project, env)
	if got.Mode != "silent" {
		t.Errorf("env should win, got %q", got.Mode)
	}
	if got.Guard.Level != "off" {
		t.Errorf("env level should win, got %q", got.Guard.Level)
	}
}

func TestMergeEnvRespectsLockdown(t *testing.T) {
	global := Config{Guard: Guard{Level: "strict", Lockdown: true}}
	env := Config{Guard: Guard{Level: "off"}}
	got, violations := MergeRespectingLockdown(global, Config{}, env)
	if got.Guard.Level != "strict" {
		t.Errorf("lockdown should block env loosening, got %q", got.Guard.Level)
	}
	if len(violations) != 1 {
		t.Errorf("expected violation, got %v", violations)
	}
}
