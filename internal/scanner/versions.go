package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"mendix-project-manager/internal/models"
)

// ScanVersions scans the version directory for installed Mendix versions
func ScanVersions(versionDir string) ([]models.Version, error) {
	var versions []models.Version

	entries, err := os.ReadDir(versionDir)
	if err != nil {
		if os.IsNotExist(err) {
			return versions, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		version, err := parseVersion(entry.Name())
		if err != nil {
			continue // Skip invalid version directories
		}

		versionPath := filepath.Join(versionDir, entry.Name())
		modelerPath := filepath.Join(versionPath, "modeler")
		studioproPath := filepath.Join(modelerPath, "studiopro.exe")
		mxPath := filepath.Join(modelerPath, "mx.exe")

		// Verify that studiopro.exe exists
		if _, err := os.Stat(studioproPath); err != nil {
			continue
		}

		version.Path = versionPath
		version.ModelerPath = modelerPath
		version.MxPath = mxPath

		versions = append(versions, version)
	}

	return versions, nil
}

// parseVersion parses a version string like "24.5.3.24567"
func parseVersion(versionStr string) (models.Version, error) {
	parts := strings.Split(versionStr, ".")
	if len(parts) != 4 {
		return models.Version{}, fmt.Errorf("invalid version format: %s", versionStr)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return models.Version{}, err
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return models.Version{}, err
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return models.Version{}, err
	}

	build, err := strconv.Atoi(parts[3])
	if err != nil {
		return models.Version{}, err
	}

	return models.Version{
		Major: major,
		Minor: minor,
		Patch: patch,
		Build: build,
	}, nil
}
