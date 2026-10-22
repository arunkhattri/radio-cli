//go:build !windows

package player

import (
	"os/exec"
	"syscall"
)

// pauseProcess sends SIGSTOP to pause the mpv process on Unix systems
func pauseProcess(cmd *exec.Cmd) error {
	if cmd != nil && cmd.Process != nil {
		return cmd.Process.Signal(syscall.SIGSTOP)
	}
	return nil
}

// resumeProcess sends SIGCONT to resume the mpv process on Unix systems
func resumeProcess(cmd *exec.Cmd) error {
	if cmd != nil && cmd.Process != nil {
		return cmd.Process.Signal(syscall.SIGCONT)
	}
	return nil
}
