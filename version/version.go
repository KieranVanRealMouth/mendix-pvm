package version

import (
	"fmt"
	"mendix-pvm/search"
	"mendix-pvm/utils"
	"path/filepath"
)

func FindModelerSubdir(dir string) (string, error) {
	p := filepath.Join(dir, "modeler")
	ok, err := utils.IsDir(p)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("no modeler directory found")
	}

	return p, nil
}

func Search(searchDirPath string, args []string) ([]string, error) {
	versions, err := search.SearchDir(searchDirPath, args)
	if err != nil {
		return nil, fmt.Errorf("An error occured while trying to search version directory\n%w", err)
	}

	filteredVersions := make([]string, 0, len(versions))
	for _, v := range versions {
		modeler, err := FindModelerSubdir(v)
		if err != nil {
			// Skip unreadable paths; optionally log with cmd.PrintErrf if desired
			continue
		}
		if modeler != "" {
			filteredVersions = append(filteredVersions, v)
		}
	}

	return filteredVersions, nil
}

func Open(versionPath string) error {
	modelerPath, err := FindModelerSubdir(versionPath)
	if err != nil {
		return err
	}

	studiopro := filepath.Join(modelerPath, "studiopro.exe")

	if err := utils.OpenFile(studiopro); err != nil {
		return err
	}

	return nil
}
