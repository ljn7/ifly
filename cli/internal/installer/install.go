// Package installer copies the embedded plugin tree into a Claude Code config
// directory and verifies the result.
package installer

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Options struct {
	Dest      string // target directory (e.g. ~/.claude/plugins/ifly)
	Overwrite bool
}

// Install walks pluginFS under the "plugin/" root and mirrors it to Dest.
// Files identified as hook scripts (see isHookExecutable) receive 0o755;
// all other files receive 0o644.
func Install(pluginFS fs.FS, o Options) error {
	if info, err := os.Stat(o.Dest); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s exists and is not a directory", o.Dest)
		}
		if !o.Overwrite {
			entries, err := os.ReadDir(o.Dest)
			if err != nil {
				return fmt.Errorf("read %s: %w", o.Dest, err)
			}
			if len(entries) > 0 {
				return fmt.Errorf("%s already populated; pass overwrite to replace", o.Dest)
			}
		}
	}
	return fs.WalkDir(pluginFS, "plugin", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "plugin" {
			return os.MkdirAll(o.Dest, 0o755)
		}
		rel := strings.TrimPrefix(path, "plugin/")
		target := filepath.Join(o.Dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(pluginFS, path, target)
	})
}

// isHookExecutable returns true if the file should be installed with 0o755.
// Two cases: any *.sh anywhere (covers lib scripts that some installs run
// directly), AND any top-level entry under plugin/hooks/ that isn't .json
// (covers the extensionless hook entry points like `guard`, `session-start`,
// and the `run-hook.cmd` polyglot wrapper).
func isHookExecutable(srcPath string) bool {
	if strings.HasSuffix(srcPath, ".sh") {
		return true
	}
	const prefix = "plugin/hooks/"
	if !strings.HasPrefix(srcPath, prefix) {
		return false
	}
	rest := strings.TrimPrefix(srcPath, prefix)
	if strings.Contains(rest, "/") {
		return false // under hooks/lib/ or any subdir
	}
	if strings.HasSuffix(rest, ".json") {
		return false // hooks.json is config, not script
	}
	return true
}

func copyFile(src fs.FS, srcPath, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := src.Open(srcPath)
	if err != nil {
		return err
	}
	defer in.Close()

	mode := os.FileMode(0o644)
	if isHookExecutable(srcPath) {
		mode = 0o755
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Chmod(mode)
}

func runtimeIsWindows() bool { return runtime.GOOS == "windows" }
