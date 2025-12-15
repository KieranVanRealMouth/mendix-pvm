package operations

import (
	"fmt"
	"mendix-pvm/internal/project"
	"mendix-pvm/internal/version"
	"os/exec"
)

func ConvertProject(ver version.Version, proj project.Project) error {
	fmt.Printf("Converting %s to version %s\n", proj.Name, ver.Name)
	mxPath, err := ver.GetExecutable()
	if err != nil {
		return err
	}

	cmd := exec.Command(mxPath, "convert", "--in-place", proj.Directory)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("conversion failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("Finished converting\n")

	return nil
}
