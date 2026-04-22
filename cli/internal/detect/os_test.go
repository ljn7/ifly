package detect

import (
	"runtime"
	"testing"
)

func TestHostReturnsRuntimeValues(t *testing.T) {
	h := Host()
	if h.OS != runtime.GOOS {
		t.Errorf("OS got %s want %s", h.OS, runtime.GOOS)
	}
	if h.Arch != runtime.GOARCH {
		t.Errorf("Arch got %s want %s", h.Arch, runtime.GOARCH)
	}
}

func TestReleaseAssetNameLinuxAmd64(t *testing.T) {
	h := HostInfo{OS: "linux", Arch: "amd64"}
	if got := h.ReleaseAsset(); got != "ifly-linux-amd64" {
		t.Errorf("got %s", got)
	}
}

func TestReleaseAssetNameWindows(t *testing.T) {
	h := HostInfo{OS: "windows", Arch: "amd64"}
	if got := h.ReleaseAsset(); got != "ifly-windows-amd64.exe" {
		t.Errorf("got %s", got)
	}
}

func TestSupportsSandboxLinuxOnly(t *testing.T) {
	if (&HostInfo{OS: "linux"}).SupportsSandbox() != true {
		t.Error("linux should support sandbox")
	}
	if (&HostInfo{OS: "darwin"}).SupportsSandbox() != false {
		t.Error("darwin should not support sandbox")
	}
}
