package project

import (
	"errors"
	"fmt"
	"mendix-pvm/search"
	"mendix-pvm/utils"
	"os"
	"path/filepath"
	"strings"
)

func FindMprAtRoot(projectPath string) (string, error) {
	info, err := os.Stat(projectPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("path does not exist: %s", projectPath)
		}
		return "", fmt.Errorf("unable to stat path %s: %w", projectPath, err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", projectPath)
	}

	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return "", fmt.Errorf("unable to read directory %s: %w", projectPath, err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(e.Name()), ".mpr") {
			return filepath.Join(projectPath, e.Name()), nil
		}
	}

	// No .mpr file found at the root
	return "", nil
}

func Search(projectDirPath string, args []string) ([]string, error) {
	projects, err := search.SearchDir(projectDirPath, args)
	if err != nil {
		return nil, fmt.Errorf("An error occured while trying to search project directory\n%w", err)
	}

	filteredProjects := make([]string, 0, len(projects))
	for _, p := range projects {

		mprPath, err := FindMprAtRoot(p)
		if err != nil {
			continue
		}
		if mprPath != "" {
			filteredProjects = append(filteredProjects, p)
		}
	}

	return filteredProjects, nil
}

func Open(projectPath string) error {
	mprPath, err := FindMprAtRoot(projectPath)
	if err != nil {
		return err
	}

	if mprPath == "" {
		return fmt.Errorf("this is not a valid project")
	}

	if err := utils.OpenFile(mprPath); err != nil {
		return err
	}

	return nil
}
