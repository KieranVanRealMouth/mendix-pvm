package cmd

import (
	"fmt"
	"os"
	"strings"

	"mendix-project-manager/internal/config"
	"mendix-project-manager/internal/models"
	"mendix-project-manager/internal/scanner"
	"mendix-project-manager/internal/search"
	"mendix-project-manager/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [type] [query...]",
	Short: "List installed Mendix versions and projects",
	Long:  `List all installed Mendix Studio Pro versions and projects, optionally filtered by a search query.`,
	Run:   runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Parse arguments
	var filterType string
	var query string

	if len(args) > 0 {
		firstArg := args[0]
		if firstArg == "versions" || firstArg == "projects" {
			filterType = firstArg
			if len(args) > 1 {
				query = strings.Join(args[1:], " ")
			}
		} else {
			query = strings.Join(args, " ")
		}
	}

	// Scan versions and projects
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

	// Filter based on type
	switch filterType {
	case "versions":
		filteredVersions := search.SearchVersions(query, versions)
		displayVersionsList(filteredVersions)
	case "projects":
		filteredProjects := search.SearchProjects(query, projects)
		displayProjectsList(filteredProjects)
	default:
		// Show both
		if query != "" {
			results := search.Search(query, versions, projects)
			displaySearchResults(results)
		} else {
			// Use table view for all results
			m := ui.NewTableModel(versions, projects)
			p := tea.NewProgram(m)
			if _, err := p.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error running UI: %v\n", err)
				os.Exit(1)
			}
		}
	}
}

func displayVersionsList(versions []models.Version) {
	if len(versions) == 0 {
		fmt.Println("No versions found.")
		return
	}

	fmt.Println("Versions:")
	for _, v := range versions {
		fmt.Printf("  - %s\n", v.String())
	}
}

func displayProjectsList(projects []models.Project) {
	if len(projects) == 0 {
		fmt.Println("No projects found.")
		return
	}

	fmt.Println("Projects:")
	for _, p := range projects {
		fmt.Printf("  - %s (%s)\n", p.Name, p.Path)
	}
}

func displaySearchResults(results []models.SearchResult) {
	if len(results) == 0 {
		fmt.Println("No results found.")
		return
	}

	for _, r := range results {
		fmt.Printf("  - %s\n", r.Display())
	}
}
