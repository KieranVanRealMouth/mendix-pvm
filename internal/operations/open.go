package operations

import (
	"fmt"
	"os/exec"
	"runtime"

	"mendix-project-manager/internal/models"
)

// OpenVersion opens a Mendix Studio Pro version
func OpenVersion(version models.Version) error {
	studioproPath := version.ModelerPath + "\\studiopro.exe"
	return openFile(studioproPath)
}

// OpenProject opens a Mendix project
func OpenProject(project models.Project) error {
	return openFile(project.MprPath)
}

// openFile opens a file with the default application
func openFile(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	return nil
}
