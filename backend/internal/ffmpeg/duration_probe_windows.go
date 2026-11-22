//go:build windows

package ffmpeg

import (
	"os/exec"
	"syscall"
)

func hideDurationProbeWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}
