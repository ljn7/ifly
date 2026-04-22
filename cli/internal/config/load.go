package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadFile reads YAML from path. A missing file returns a zero Config with nil error.
// An invalid file returns an error.
func LoadFile(path string) (Config, error) {
	// An empty path means no project file is configured (CLAUDE_PROJECT_DIR unset).
	// Treat that the same as "file not present" and return a zero config.
	if path == "" {
		return Config{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return c, nil
}

// LoadBytes parses a raw YAML blob, used for embedded defaults.yaml.
func LoadBytes(data []byte) (Config, error) {
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return Config{}, fmt.Errorf("parse bytes: %w", err)
	}
	return c, nil
}

// LoadEnv pulls IFLY_MODE and IFLY_GUARD into a sparse Config.
func LoadEnv() Config {
	c := Config{}
	if v := os.Getenv("IFLY_MODE"); v != "" {
		c.Mode = v
	}
	if v := os.Getenv("IFLY_GUARD"); v != "" {
		c.Guard.Level = v
	}
	return c
}
