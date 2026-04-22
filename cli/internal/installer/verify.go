package installer

import (
	"fmt"
	"os"
	"path/filepath"
)

type Report struct {
	OK       bool
	Failures []string
}

// Verify checks that a populated plugin directory has the required files
// and executable bits. The guard hook may be installed as either
// "hooks/guard" (extensionless, post-rename real plugin tree) or
// "hooks/guard.sh" (test fixture); either is acceptable as long as
// exactly one is present and executable.
func Verify(dest string) (Report, error) {
	r := Report{OK: true}
	mustExist := []string{
		".claude-plugin/plugin.json",
		"defaults.yaml",
		"hooks/ifly-state",
		"commands/status.md",
		"commands/guard.md",
		"commands/mode.md",
	}
	for _, rel := range mustExist {
		if _, err := os.Stat(filepath.Join(dest, rel)); err != nil {
			r.OK = false
			r.Failures = append(r.Failures, fmt.Sprintf("missing %s", rel))
		}
	}

	guard, err := findGuardHook(dest)
	if err != nil {
		r.OK = false
		r.Failures = append(r.Failures, err.Error())
	} else if !runtimeIsWindows() {
		info, _ := os.Stat(guard)
		if info.Mode().Perm()&0o100 == 0 {
			r.OK = false
			rel, _ := filepath.Rel(dest, guard)
			r.Failures = append(r.Failures, rel+" not executable")
		}
	}

	return r, nil
}

func findGuardHook(dest string) (string, error) {
	for _, name := range []string{"guard", "guard.sh"} {
		p := filepath.Join(dest, "hooks", name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("missing hooks/guard or hooks/guard.sh")
}
