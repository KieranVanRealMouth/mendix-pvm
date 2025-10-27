package cmd

import (
	"fmt"
	"os"

	"mendix-project-manager/internal/models"
	"mendix-project-manager/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func selectVersionFromList(versions []models.Version) *models.Version {
	searchResults := make([]models.SearchResult, len(versions))
	for i := range versions {
		searchResults[i] = models.SearchResult{
			Type:    "version",
			Version: &versions[i],
		}
	}

	m := ui.NewListModel(searchResults, "Select version:")
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running UI: %v\n", err)
		return nil
	}

	listModel := finalModel.(ui.ListModel)
	if listModel.Choice() != nil && listModel.Choice().Version != nil {
		return listModel.Choice().Version
	}
	return nil
}

func selectProjectFromList(projects []models.Project) *models.Project {
	searchResults := make([]models.SearchResult, len(projects))
	for i := range projects {
		searchResults[i] = models.SearchResult{
			Type:    "project",
			Project: &projects[i],
		}
	}

	m := ui.NewListModel(searchResults, "Select project:")
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running UI: %v\n", err)
		return nil
	}

	listModel := finalModel.(ui.ListModel)
	if listModel.Choice() != nil && listModel.Choice().Project != nil {
		return listModel.Choice().Project
	}
	return nil
}
