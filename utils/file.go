package utils

import (
	"fmt"
	"os/exec"
	"runtime"
)

func OpenFile(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// 'start' needs to run inside cmd.exe and requires an empty title arg
		cmd = exec.Command("cmd", "/c", "start", "", path)

	case "darwin":
		cmd = exec.Command("open", path)

	default: // Linux, BSD, etc.
		cmd = exec.Command("xdg-open", path)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	return nil
}
