// NEW FILE: pkg/convert/convert.go
package convert

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

func Convert(versionPath, projectPath string) (int, error) {
	mxExePath := filepath.Join(versionPath, "modeler", "mx.exe")
	cmd := exec.Command(mxExePath, "convert", "--in-place", projectPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err == nil {
		return 0, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), err
	}

	return -1, err
}
