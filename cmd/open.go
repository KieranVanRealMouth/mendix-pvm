package cmd

import (
	"fmt"
	"os"
	"strings"

	"mendix-project-manager/internal/config"
	"mendix-project-manager/internal/models"
	"mendix-project-manager/internal/operations"
	"mendix-project-manager/internal/scanner"
	"mendix-project-manager/internal/search"
	"mendix-project-manager/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	openVersion bool
	openProject bool
)

var openCmd = &cobra.Command{
	Use:   "open [query...]",
	Short: "Open a Mendix version or project",
	Long:  `Open a Mendix Studio Pro version or project. If multiple matches are found, you can select from a list.`,
	Run:   runOpen,
}

func init() {
	openCmd.Flags().BoolVarP(&openVersion, "version", "v", false, "Only consider versions")
	openCmd.Flags().BoolVarP(&openProject, "project", "p", false, "Only consider projects")
	rootCmd.AddCommand(openCmd)
}

func runOpen(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	query := strings.Join(args, " ")

	versions, err := scanner.ScanVersions(cfg.VersionDirectory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning versions: %v\n", err)
		os.Exit(1)
	}

	projects, err := scanner.ScanProjects(cfg.ProjectDirectories)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning projects: %v\n", err)
		os.Exit(1)
	}

	if openVersion {
		filteredVersions := search.SearchVersions(query, versions)
		if len(filteredVersions) == 0 {
			fmt.Println("No versions found.")
			return
		} else if len(filteredVersions) == 1 {
			err := operations.OpenVersion(filteredVersions[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening version: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Opened version %s\n", filteredVersions[0].String())
			return
		} else {
			selectedVersion := selectVersionFromList(filteredVersions)
			if selectedVersion != nil {
				err := operations.OpenVersion(*selectedVersion)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error opening version: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("Opened version %s\n", selectedVersion.String())
			}
			return
		}
	} else if openProject {
		filteredProjects := search.SearchProjects(query, projects)
		if len(filteredProjects) == 0 {
			fmt.Println("No projects found.")
			return
		} else if len(filteredProjects) == 1 {
			err := operations.OpenProject(filteredProjects[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening project: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Opened project %s\n", filteredProjects[0].Name)
			return
		} else {
			selectedProject := selectProjectFromList(filteredProjects)
			if selectedProject != nil {
				err := operations.OpenProject(*selectedProject)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error opening project: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("Opened project %s\n", selectedProject.Name)
			}
			return
		}
	} else {
		searchResults := search.Search(query, versions, projects)
		if len(searchResults) == 0 {
			fmt.Println("No results found.")
			return
		} else if len(searchResults) == 1 {
			openSearchResult(&searchResults[0])
			return
		} else {
			// Show list for selection
			m := ui.NewListModel(searchResults, "Select item to open:")
			p := tea.NewProgram(m)
			finalModel, err := p.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error running UI: %v\n", err)
				os.Exit(1)
			}

			listModel := finalModel.(ui.ListModel)
			if listModel.Choice() != nil {
				openSearchResult(listModel.Choice())
			}
			return
		}
	}
}

func openSearchResult(result *models.SearchResult) {
	if result.Type == "version" && result.Version != nil {
		err := operations.OpenVersion(*result.Version)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening version: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Opened version %s\n", result.Version.String())
	} else if result.Type == "project" && result.Project != nil {
		err := operations.OpenProject(*result.Project)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening project: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Opened project %s\n", result.Project.Name)
	}
}
