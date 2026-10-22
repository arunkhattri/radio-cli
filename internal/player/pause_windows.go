//go:build windows

package player

import (
	"os/exec"
)

// pauseProcess on Windows (Signals are not natively supported like POSIX)
func pauseProcess(cmd *exec.Cmd) error {
	// Windows process suspension can be handled via Windows API or fallback
	return nil
}

// resumeProcess on Windows
func resumeProcess(cmd *exec.Cmd) error {
	return nil
}
