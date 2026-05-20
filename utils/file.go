package utils

import (
	"fmt"
	"os/exec"
)

func OpenFile(path string) error {
	var cmd *exec.Cmd

	switch Platform() {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "wsl":
		if err := checkWslview(); err != nil {
			return err
		}
		cmd = exec.Command("wslview", path)
	default: // linux
		cmd = exec.Command("xdg-open", path)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	return nil
}

func checkWslview() error {
	if _, err := exec.LookPath("wslview"); err != nil {
		return fmt.Errorf("wslview is not installed — install wslu with: sudo apt install wslu")
	}
	return nil
}
