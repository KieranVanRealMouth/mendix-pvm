package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"mendix-project-manager/internal/models"
)

// ScanProjects scans the configured directories for Mendix projects
func ScanProjects(projectDirs []string) ([]models.Project, error) {
	var projects []models.Project

	for _, dir := range projectDirs {
		foundProjects, err := scanDirectory(dir)
		if err != nil {
			// Continue with other directories even if one fails
			continue
		}
		projects = append(projects, foundProjects...)
	}

	return projects, nil
}

// scanDirectory scans a single directory for Mendix projects
func scanDirectory(dir string) ([]models.Project, error) {
	var projects []models.Project

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectPath := filepath.Join(dir, entry.Name())
		mprPath, found := findMprFile(projectPath)
		if found {
			projects = append(projects, models.Project{
				Name:    entry.Name(),
				Path:    projectPath,
				MprPath: mprPath,
			})
		}
	}

	return projects, nil
}

// findMprFile looks for a .mpr file in the project directory
func findMprFile(projectPath string) (string, bool) {
	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return "", false
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".mpr") {
			return filepath.Join(projectPath, entry.Name()), true
		}
	}

	return "", false
}
