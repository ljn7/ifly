//go:build linux

package sandbox

import (
	"bytes"
	"os/exec"
	"testing"
)

func TestAvailableReflectsBwrapOrUnshare(t *testing.T) {
	haveBwrap := exec.Command("bwrap", "--version").Run() == nil
	haveUnshare := exec.Command("unshare", "--version").Run() == nil
	want := haveBwrap || haveUnshare
	if got := Available(); got != want {
		t.Errorf("Available=%v want %v", got, want)
	}
}

func TestBuildBwrapArgs(t *testing.T) {
	args := buildBwrapArgs("/home/me/proj", []string{"claude", "--version"})
	joined := " " + stringsJoin(args, " ") + " "

	mustContain := []string{
		" --ro-bind /usr /usr ",
		" --bind /home/me/proj /home/me/proj ",
		" --tmpfs /tmp ",
		" --unshare-pid ",
		" --die-with-parent ",
		" claude ",
		" --version",
	}
	for _, m := range mustContain {
		if !bytes.Contains([]byte(joined), []byte(m)) {
			t.Errorf("missing %q in args: %s", m, joined)
		}
	}
}

// local joiner so tests don't drag in strings import into build file
func stringsJoin(xs []string, sep string) string {
	out := ""
	for i, x := range xs {
		if i > 0 {
			out += sep
		}
		out += x
	}
	return out
}
