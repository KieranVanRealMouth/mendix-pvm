package operations

import (
	"fmt"
	"os/exec"

	"mendix-project-manager/internal/models"
)

// ConvertProject converts a project to a specific version
func ConvertProject(project models.Project, version models.Version) error {
	mxPath := version.MxPath

	cmd := exec.Command(mxPath, "convert", "--in-place", project.MprPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("conversion failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}
