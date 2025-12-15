package version

import (
	"fmt"
	"os"
	"path/filepath"
)

type Version struct {
	Name      string
	Directory string
	StudioPro string
}

func CreateVersion(p string) (Version, error) {
	modelerPath := "modeler"
	studioProFile := "studiopro.exe"
	absolutePath, err := filepath.Abs(p)
	if err != nil {
		return Version{}, fmt.Errorf("failed to get absolute path for: %s with error: %w", p, err)
	}

	if os.IsNotExist(err) {
		return Version{}, fmt.Errorf("directory does not exist: %s with error: %w", p, err)
	}

	studioProPath := filepath.Join(p, modelerPath, studioProFile)

	if _, err := os.Stat(studioProPath); os.IsNotExist(err) {
		return Version{}, err
	}

	return Version{Name: filepath.Base(p), Directory: absolutePath, StudioPro: studioProPath}, nil
}
