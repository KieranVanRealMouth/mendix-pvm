package cmd

import (
	"fmt"
	"os"

	"mendix-project-manager/internal/config"
	"mendix-project-manager/internal/models"
	"mendix-project-manager/internal/operations"
	"mendix-project-manager/internal/scanner"
	"mendix-project-manager/internal/search"

	"github.com/spf13/cobra"
)

var (
	convertVersion string
	convertProject string
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert a project to a different Mendix version",
	Long:  `Convert a Mendix project to a different Studio Pro version using the mx.exe tool.`,
	Run:   runConvert,
}

func init() {
	convertCmd.Flags().StringVarP(&convertVersion, "version", "v", "", "Version to convert to")
	convertCmd.Flags().StringVarP(&convertProject, "project", "p", "", "Project to convert")
	rootCmd.AddCommand(convertCmd)
}

func runConvert(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

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

	var selectedVersion *models.Version
	var selectedProject *models.Project

	// Select version
	if convertVersion != "" {
		filteredVersions := search.SearchVersions(convertVersion, versions)
		if len(filteredVersions) == 0 {
			fmt.Println("No matching version found.")
			return
		} else if len(filteredVersions) == 1 {
			selectedVersion = &filteredVersions[0]
		} else {
			selectedVersion = selectVersionFromList(filteredVersions)
		}
	} else {
		selectedVersion = selectVersionFromList(versions)
	}

	if selectedVersion == nil {
		return
	}

	// Select project
	if convertProject != "" {
		filteredProjects := search.SearchProjects(convertProject, projects)
		if len(filteredProjects) == 0 {
			fmt.Println("No matching project found.")
			return
		} else if len(filteredProjects) == 1 {
			selectedProject = &filteredProjects[0]
		} else {
			selectedProject = selectProjectFromList(filteredProjects)
		}
	} else {
		selectedProject = selectProjectFromList(projects)
	}

	if selectedProject == nil {
		return
	}

	// Perform conversion
	fmt.Printf("Converting project '%s' to version %s...\n", selectedProject.Name, selectedVersion.String())

	err = operations.ConvertProject(*selectedProject, *selectedVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting project: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Conversion completed successfully!")
}
