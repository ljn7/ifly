//go:build !linux

package sandbox

import (
	"errors"
	"io"
)

func Available() bool { return false }

func Run(_ io.Writer, _ io.Writer, _ []string, _ string) error {
	return errors.New("sandbox requires Linux namespaces; on macOS/Windows use `/ifly:guard strict`")
}
