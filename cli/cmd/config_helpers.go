package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/ljn7/ifly/cli/internal/config"
)

// filepathDir is a tiny indirection so tests on windows don't fight backslashes.
func filepathDir(p string) string { return filepath.Dir(p) }

func validateSet(key, val string) error {
	switch key {
	case "mode":
		return config.ValidateMode(val)
	case "guard.level":
		return config.ValidateLevel(val)
	case "guard.lockdown":
		if val != "true" && val != "false" {
			return fmt.Errorf("guard.lockdown must be true or false")
		}
	}
	return nil
}

func setDotted(c *config.Config, key, val string) error {
	switch key {
	case "mode":
		c.Mode = val
	case "guard.level":
		c.Guard.Level = val
	case "guard.lockdown":
		b, _ := strconv.ParseBool(val)
		c.Guard.Lockdown = b
	case "telemetry.easter_egg":
		b, _ := strconv.ParseBool(val)
		c.Telemetry.EasterEgg = b
	default:
		return fmt.Errorf("unknown key %q; supported: mode, guard.level, guard.lockdown, telemetry.easter_egg", key)
	}
	return nil
}

func getDotted(c config.Config, key string) (string, error) {
	switch key {
	case "mode":
		return c.Mode, nil
	case "guard.level":
		return c.Guard.Level, nil
	case "guard.lockdown":
		return fmt.Sprintf("%v", c.Guard.Lockdown), nil
	case "telemetry.easter_egg":
		return fmt.Sprintf("%v", c.Telemetry.EasterEgg), nil
	}
	return "", fmt.Errorf("unknown key %q", key)
}
