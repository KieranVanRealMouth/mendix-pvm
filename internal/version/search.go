package version

import (
	"mendix-pvm/internal/config"
	"mendix-pvm/internal/search"
	"os"
	"path/filepath"
)

func Search(args []string) ([]Version, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	results, err := search.Search(cfg.VersionDirectory, args)
	if err != nil {
		return nil, err
	}

	var returnValue []Version

	for _, result := range results {
		studioProPath := filepath.Join(result.Path, "modeler", "studiopro.exe")
		if _, err := os.Stat(studioProPath); os.IsNotExist(err) {
			continue
		}

		proj, err := CreateVersion(result.Path)
		if err != nil {
			return nil, err
		}

		returnValue = append(returnValue, proj)
	}

	return returnValue, nil
}
