package detect

import "runtime"

type HostInfo struct {
	OS   string
	Arch string
}

func Host() HostInfo {
	return HostInfo{OS: runtime.GOOS, Arch: runtime.GOARCH}
}

func (h HostInfo) ReleaseAsset() string {
	name := "ifly-" + h.OS + "-" + h.Arch
	if h.OS == "windows" {
		name += ".exe"
	}
	return name
}

func (h HostInfo) SupportsSandbox() bool {
	return h.OS == "linux"
}
