package project

import (
	"fmt"
	"os"
	"path/filepath"
)

type Project struct {
	Name      string
	Directory string
	Mpr       string
}

func CreateProject(p string) (Project, error) {
	ext := ".mpr"
	absolutePath, err := filepath.Abs(p)
	if err != nil {
		return Project{}, fmt.Errorf("failed to get absolute path for: %s with error: %w", p, err)

	}

	entries, err := os.ReadDir(absolutePath)
	if os.IsNotExist(err) {

		return Project{}, fmt.Errorf("directory does not exist: %s with error: %w", p, err)
	}

	var mprPath string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) == ext {
			mprPath = entry.Name()
		}
	}

	return Project{Name: filepath.Base(p), Directory: absolutePath, Mpr: mprPath}, nil
}
