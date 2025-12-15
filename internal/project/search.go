package project

import (
	"mendix-pvm/internal/config"
	"mendix-pvm/internal/search"
	"os"
	"strings"
)

func Search(args []string) ([]Project, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	results, err := search.Search(cfg.ProjectDirectory, args)
	if err != nil {
		return nil, err
	}

	var returnValue []Project

	for _, result := range results {
		entries, err := os.ReadDir(result.Path)
		if err != nil {
			return nil, err
		}

		isProject := false

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			if !strings.HasSuffix(entry.Name(), ".user.json") {
				continue
			}

			isProject = true
		}

		if isProject {
			proj, err := CreateProject(result.Path)
			if err != nil {
				return nil, err
			}

			returnValue = append(returnValue, proj)
		}
	}

	return returnValue, nil
}
