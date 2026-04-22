// Package state reads and writes IFLy's session state file.
package state

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type State struct {
	Version              int       `yaml:"version"`
	ActiveMode           string    `yaml:"active_mode,omitempty"`
	SessionGuardOverride string    `yaml:"session_guard_override,omitempty"`
	UpdatedAt            time.Time `yaml:"updated_at,omitempty"`
}

func Merge(base, overlay State) State {
	out := base
	if overlay.Version != 0 {
		out.Version = overlay.Version
	}
	if overlay.ActiveMode != "" {
		out.ActiveMode = overlay.ActiveMode
	}
	if overlay.SessionGuardOverride != "" {
		out.SessionGuardOverride = overlay.SessionGuardOverride
	}
	if !overlay.UpdatedAt.IsZero() {
		out.UpdatedAt = overlay.UpdatedAt
	}
	if out.Version == 0 {
		out.Version = 1
	}
	return out
}

func Load(path string) (State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return State{Version: 1}, nil
		}
		return State{}, fmt.Errorf("read %s: %w", path, err)
	}
	var s State
	if err := yaml.Unmarshal(data, &s); err != nil {
		return State{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return s, nil
}

func Save(path string, s State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	// Preserve caller-set timestamp; stamp now only when zero. This keeps
	// the round-trip test deterministic and lets callers control the value
	// (e.g., for tests or replay) while still defaulting to "now" in the
	// normal path.
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = time.Now().UTC().Truncate(time.Second)
	}
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}
