// Package paths resolves OS-appropriate locations for IFLy config and state.
package paths

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// GlobalConfigDir returns the directory that holds config.yaml and state.yaml.
//
//	linux/bsd:  $XDG_CONFIG_HOME/ifly  (fallback $HOME/.config/ifly)
//	darwin:     $HOME/Library/Application Support/ifly
//	windows:    %APPDATA%/ifly
func GlobalConfigDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		if v := os.Getenv("APPDATA"); v != "" {
			return filepath.Join(v, "ifly"), nil
		}
		return "", errors.New("APPDATA not set")
	case "darwin":
		h, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(h, "Library", "Application Support", "ifly"), nil
	default:
		if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
			return filepath.Join(v, "ifly"), nil
		}
		h, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(h, ".config", "ifly"), nil
	}
}

func GlobalConfigFile() (string, error) {
	d, err := GlobalConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "config.yaml"), nil
}

func StateFile() (string, error) {
	d, err := GlobalConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "state.yaml"), nil
}

// ProjectRoot returns CLAUDE_PROJECT_DIR when Claude provides it, otherwise
// the current working directory for CLI use outside Claude.
func ProjectRoot() (string, error) {
	if v := os.Getenv("CLAUDE_PROJECT_DIR"); v != "" {
		return v, nil
	}
	return os.Getwd()
}

func ProjectStateFile() (string, error) {
	root, err := ProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, ".ifly-state.yaml"), nil
}

// ProjectConfigFile returns $CLAUDE_PROJECT_DIR/.ifly.yaml, or "" if unset.
func ProjectConfigFile() string {
	if v := os.Getenv("CLAUDE_PROJECT_DIR"); v != "" {
		return filepath.Join(v, ".ifly.yaml")
	}
	return ""
}

// EnsureDir creates the directory if missing, with 0o755.
func EnsureDir(p string) error {
	return os.MkdirAll(p, 0o755)
}
