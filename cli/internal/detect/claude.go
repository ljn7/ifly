package detect

import (
	"fmt"
	"os"
	"path/filepath"
)

// ClaudeConfigDir walks per-OS candidate paths and returns the first that exists.
// Returns an error if no candidate exists (installer may still create the
// preferred candidate when running --overwrite).
func ClaudeConfigDir(h HostInfo) (string, error) {
	candidates, err := claudeCandidates(h)
	if err != nil {
		return "", err
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("no Claude config dir found; tried %v", candidates)
}

// ClaudeConfigDirPreferred returns the first candidate (creation target) even
// if it does not exist. Callers are responsible for mkdir.
func ClaudeConfigDirPreferred(h HostInfo) (string, error) {
	candidates, err := claudeCandidates(h)
	if err != nil {
		return "", err
	}
	return candidates[0], nil
}

// ClaudeConfigCandidates returns the per-OS Claude config directory candidates
// in priority order. The installer uses the first path as its creation target.
func ClaudeConfigCandidates(h HostInfo) ([]string, error) {
	return claudeCandidates(h)
}

func claudeCandidates(h HostInfo) ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	switch h.OS {
	case "windows":
		appdata := os.Getenv("APPDATA")
		list := []string{}
		if appdata != "" {
			list = append(list, filepath.Join(appdata, "claude"))
		}
		list = append(list, filepath.Join(home, ".claude"))
		return list, nil
	case "darwin":
		return []string{
			filepath.Join(home, "Library", "Application Support", "claude"),
			filepath.Join(home, ".claude"),
		}, nil
	default:
		list := []string{}
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			list = append(list, filepath.Join(xdg, "claude"))
		}
		list = append(list,
			filepath.Join(home, ".config", "claude"),
			filepath.Join(home, ".claude"),
		)
		return list, nil
	}
}

// ClaudeBinaryPresent returns true if `claude` is resolvable on PATH.
func ClaudeBinaryPresent() bool {
	_, err := execLookPath("claude")
	return err == nil
}

// BashLikelyAvailable is a best-effort check for a bash shell. On Linux/macOS
// this is almost always true; on Windows it signals Git Bash or WSL is
// installed and the plugin's hooks can actually run.
func BashLikelyAvailable() bool {
	_, err := execLookPath("bash")
	return err == nil
}

// Overridable for tests.
var execLookPath = defaultLookPath
