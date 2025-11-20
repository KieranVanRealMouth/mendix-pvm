package cmd

import (
	"fmt"
	"mendix-project-manager/internal/config"
	"mendix-project-manager/internal/models"
	"mendix-project-manager/internal/operations"
	"mendix-project-manager/internal/scanner"
	"mendix-project-manager/internal/search"
	"mendix-project-manager/internal/ui"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var outputFile string

var unmergedCmd = &cobra.Command{
	Use:   "unmerged [query] --output /path/to/output/file.txt",
	Short: "Shows unmerged commits for the project",
	Long:  "Fetches from git and shows all unmerged commits and their branches relative to the specified project and the active branch",
	Run:   runUnmerged,
}

func init() {
	unmergedCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path")
	rootCmd.AddCommand(unmergedCmd)
}

func runUnmerged(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	projects, err := scanner.ScanProjects(cfg.ProjectDirectories)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning projects: %v\n", err)
		os.Exit(1)
	}

	query := strings.Join(args, " ")

	searchResults := search.Search(query, nil, projects)
	if len(searchResults) == 0 {
		fmt.Println("No results found.")
		return
	} else if len(searchResults) == 1 {
		unmergedSearchResult(&searchResults[0])
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
			unmergedSearchResult(listModel.Choice())
		}
		return
	}
}

func unmergedSearchResult(result *models.SearchResult) {
	if result.Type == "version" {
		fmt.Printf("failed to get unmerged branches because a version was chosen")
		return
	}

	if result.Project == nil {
		fmt.Printf("failed to get unmerged branches because project is nil")
	}

	currentBranch, err := operations.GetCurrentBranch(result.Project.Path)
	if err != nil {
		fmt.Printf("Error getting current branch: %v\n", err)
		os.Exit(1)
	}

	unmergedBranches, err := operations.GetUnmergedBranches(result.Project.Path)
	if err != nil {
		fmt.Printf("error getting unmerged branches: %v\n", err)
		os.Exit(1)
	}

	output := operations.BuildOutput(result.Project.Path, currentBranch, unmergedBranches)

	filename := outputFile
	if filename == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error getting home directory: %v\n", err)
			os.Exit(1)
		}
		filename = filepath.Join(homeDir, fmt.Sprintf("%s-unmerged-branches.txt", currentBranch))
	}

	if err := os.WriteFile(filename, []byte(output), 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}
}
