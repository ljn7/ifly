// Package config models IFLy's YAML config shared with the bash hook.
package config

import "fmt"

type Config struct {
	Version   int       `yaml:"version"`
	Mode      string    `yaml:"mode,omitempty"`
	Guard     Guard     `yaml:"guard,omitempty"`
	Telemetry Telemetry `yaml:"telemetry,omitempty"`
}

type Guard struct {
	Level           string   `yaml:"level,omitempty"`
	Lockdown        bool     `yaml:"lockdown,omitempty"`
	Tools           Tools    `yaml:"tools,omitempty"`
	AdditionalDirs  []string `yaml:"additional_dirs,omitempty"`
	BlockedCommands []string `yaml:"blocked_commands,omitempty"`
	AllowedNetwork  []string `yaml:"allowed_network,omitempty"`
	SensitivePaths  []string `yaml:"sensitive_paths,omitempty"`
}

// Tools uses *bool so "not set" (nil) is distinguishable from "explicitly false".
type Tools struct {
	Bash         *bool `yaml:"bash,omitempty"`
	Edit         *bool `yaml:"edit,omitempty"`
	Write        *bool `yaml:"write,omitempty"`
	MultiEdit    *bool `yaml:"multi_edit,omitempty"`
	NotebookEdit *bool `yaml:"notebook_edit,omitempty"`
	Read         *bool `yaml:"read,omitempty"`
	Glob         *bool `yaml:"glob,omitempty"`
	Grep         *bool `yaml:"grep,omitempty"`
	WebFetch     *bool `yaml:"web_fetch,omitempty"`
	WebSearch    *bool `yaml:"web_search,omitempty"`
}

type Telemetry struct {
	EasterEgg bool `yaml:"easter_egg"`
}

var validModes = map[string]struct{}{
	"silent": {}, "minimal": {}, "normal": {}, "verbose": {}, "caveman": {},
}

func ValidateMode(m string) error {
	if _, ok := validModes[m]; !ok {
		return fmt.Errorf("invalid mode %q; want silent|minimal|normal|verbose|caveman", m)
	}
	return nil
}

var levelRanks = map[string]int{"strict": 3, "project": 2, "open": 1, "off": 0}

func ValidateLevel(l string) error {
	if _, ok := levelRanks[l]; !ok {
		return fmt.Errorf("invalid guard level %q; want strict|project|open|off", l)
	}
	return nil
}

// LevelRank returns the ordering used by lockdown checks: higher = stricter.
// Returns -1 for unknown levels.
func LevelRank(l string) int {
	if r, ok := levelRanks[l]; ok {
		return r
	}
	return -1
}
