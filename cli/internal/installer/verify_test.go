package installer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyPassesOnFullInstall(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "claude")
	if err := Install(fakePluginFS(), Options{Dest: dst}); err != nil {
		t.Fatal(err)
	}
	report, err := Verify(dst)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !report.OK {
		t.Errorf("expected OK, got failures: %v", report.Failures)
	}
}

func TestVerifyFailsWhenManifestMissing(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "claude")
	if err := Install(fakePluginFS(), Options{Dest: dst}); err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(filepath.Join(dst, ".claude-plugin", "plugin.json"))
	report, err := Verify(dst)
	if err != nil {
		t.Fatal(err)
	}
	if report.OK {
		t.Error("expected failure when plugin.json missing")
	}
}

func TestVerifyFailsWhenGuardNotExecutable(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("executable bit not tracked on windows")
	}
	dst := filepath.Join(t.TempDir(), "claude")
	if err := Install(fakePluginFS(), Options{Dest: dst}); err != nil {
		t.Fatal(err)
	}
	_ = os.Chmod(filepath.Join(dst, "hooks", "guard.sh"), 0o644)
	report, err := Verify(dst)
	if err != nil {
		t.Fatal(err)
	}
	if report.OK {
		t.Error("expected failure when guard.sh not executable")
	}
}
