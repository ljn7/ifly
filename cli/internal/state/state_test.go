package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadMissingReturnsZero(t *testing.T) {
	p := filepath.Join(t.TempDir(), "state.yaml")
	s, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if s.ActiveMode != "" {
		t.Errorf("expected empty mode, got %q", s.ActiveMode)
	}
}

func TestRoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "state.yaml")
	in := State{
		Version:              1,
		ActiveMode:           "silent",
		SessionGuardOverride: "open",
		UpdatedAt:            time.Now().UTC().Truncate(time.Second),
	}
	if err := Save(p, in); err != nil {
		t.Fatal(err)
	}
	out, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if out.ActiveMode != in.ActiveMode ||
		out.SessionGuardOverride != in.SessionGuardOverride ||
		!out.UpdatedAt.Equal(in.UpdatedAt) {
		t.Errorf("roundtrip mismatch: got %+v want %+v", out, in)
	}
}

func TestSaveCreatesParentDir(t *testing.T) {
	p := filepath.Join(t.TempDir(), "nested", "dirs", "state.yaml")
	if err := Save(p, State{Version: 1, ActiveMode: "minimal"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Errorf("stat after save: %v", err)
	}
}

func TestMergeOverlaysNonEmptyValues(t *testing.T) {
	base := State{Version: 1, ActiveMode: "normal", SessionGuardOverride: "open"}
	overlay := State{Version: 1, ActiveMode: "silent"}
	got := Merge(base, overlay)
	if got.ActiveMode != "silent" || got.SessionGuardOverride != "open" {
		t.Fatalf("merge mismatch: %+v", got)
	}
}
