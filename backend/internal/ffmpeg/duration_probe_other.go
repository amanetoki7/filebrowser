//go:build !windows

package ffmpeg

import "os/exec"

func hideDurationProbeWindow(cmd *exec.Cmd) {}
