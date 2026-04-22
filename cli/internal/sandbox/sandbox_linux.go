//go:build linux

// Package sandbox wraps `claude` with a filesystem namespace on Linux.
// Prefers bubblewrap (bwrap); falls back to unshare. Non-Linux builds
// use the stub in sandbox_other.go.
package sandbox

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Available reports whether bwrap or unshare is resolvable on PATH.
func Available() bool {
	if _, err := exec.LookPath("bwrap"); err == nil {
		return true
	}
	if _, err := exec.LookPath("unshare"); err == nil {
		return true
	}
	return false
}

// Run starts `claude` (or the provided argv) inside a sandbox. projectDir
// is read-write-bound so Claude can edit files in it; the rest of the FS is
// read-only (bwrap) or namespace-isolated (unshare fallback).
func Run(stdout, stderr io.Writer, argv []string, projectDir string) error {
	if len(argv) == 0 {
		return errors.New("sandbox: empty argv")
	}
	if projectDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		projectDir = wd
	}

	if _, err := exec.LookPath("bwrap"); err == nil {
		args := buildBwrapArgs(projectDir, argv)
		cmd := exec.Command("bwrap", args...)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	}
	if _, err := exec.LookPath("unshare"); err == nil {
		fmt.Fprintln(stderr, "sandbox: bwrap not found; falling back to unshare (weaker isolation)")
		args := buildUnshareArgs(argv)
		cmd := exec.Command("unshare", args...)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	}
	return errors.New("sandbox: neither bwrap nor unshare on PATH")
}

func buildBwrapArgs(projectDir string, argv []string) []string {
	args := []string{
		"--ro-bind", "/usr", "/usr",
		"--ro-bind", "/etc", "/etc",
		"--ro-bind", "/bin", "/bin",
		"--ro-bind", "/lib", "/lib",
		"--ro-bind", "/lib64", "/lib64",
		"--tmpfs", "/tmp",
		"--bind", projectDir, projectDir,
		"--dev", "/dev",
		"--proc", "/proc",
		"--unshare-pid",
		"--unshare-ipc",
		"--unshare-uts",
		"--unshare-cgroup",
		"--die-with-parent",
	}
	return append(args, argv...)
}

func buildUnshareArgs(argv []string) []string {
	return append([]string{"--mount", "--pid", "--fork"}, argv...)
}
