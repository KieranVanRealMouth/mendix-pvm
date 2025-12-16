package operations

import (
	"fmt"
	"mendix-pvm/internal/project"
	"mendix-pvm/internal/version"
	"os/exec"
	"sync"
)

var outputMutex sync.Mutex

func ConvertProject(ver version.Version, proj project.Project) error {
	outputMutex.Lock()
	fmt.Printf("Converting %s to version %s\n", proj.Name, ver.Name)
	outputMutex.Unlock()

	mxPath, err := ver.GetExecutable()
	if err != nil {
		outputMutex.Lock()
		fmt.Printf("Failed to get project execute for %s\n", proj.Name)
		outputMutex.Unlock()
		return err
	}

	cmd := exec.Command(mxPath, "convert", "--in-place", proj.Directory)

	output, err := cmd.CombinedOutput()
	if err != nil {
		outputMutex.Lock()
		fmt.Printf("Failed to convert %s\n", proj.Name)
		outputMutex.Unlock()
		return fmt.Errorf("conversion failed: %w\nOutput: %s", err, string(output))
	}

	outputMutex.Lock()
	fmt.Printf("Finished converting %s\n", proj.Name)
	outputMutex.Unlock()

	return nil
}
