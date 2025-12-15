package cmd

import (
	"fmt"
	"mendix-pvm/internal/config"
	"mendix-pvm/internal/operations"
	"mendix-pvm/internal/project"
	"mendix-pvm/internal/search"
	"mendix-pvm/internal/version"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	projectFlag bool
	versionFlag bool
)

var openCmd = &cobra.Command{
	Use:   "open [query] [flags]",
	Short: "Open the specified version or project",
	Long:  "Open the specified version or project using the default program. When nothing is specified the program shows an error message.",
	RunE:  runOpen,
}

func init() {
	openCmd.Flags().BoolVar(&projectFlag, "project", false, "Only search for projects.")
	openCmd.Flags().BoolVar(&versionFlag, "version", false, "Only search for versions.")
	rootCmd.AddCommand(openCmd)
}

func runOpen(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var directories []string

	if projectFlag {
		directories = append(directories, cfg.ProjectDirectory)
	}

	if versionFlag {
		directories = append(directories, cfg.VersionDirectory)
	}

	if directories == nil {
		directories = append(directories, cfg.ProjectDirectory, cfg.VersionDirectory)
	}

	var results []search.SearchResult
	for _, dir := range directories {
		dirResult, err := search.Search(dir, args)
		if err != nil {
			return err
		}

		results = append(results, dirResult...)
	}

	x := len(results)

	if x <= 0 {
		return fmt.Errorf("no results found for search query")
	}

	if x > 1 {
		fmt.Printf("Found %v results\n", x)
		for _, res := range results {
			fmt.Printf("- %s\n", res.Name)
		}
		return nil
	}

	result := results[0]

	if strings.HasPrefix(result.Path, cfg.ProjectDirectory) {
		proj, err := project.CreateProject(result.Path)
		if err != nil {
			return err
		}

		operations.OpenFile(filepath.Join(proj.Directory, proj.Mpr))
	} else if strings.HasPrefix(result.Path, cfg.VersionDirectory) {
		ver, err := version.CreateVersion(result.Path)
		if err != nil {
			return err
		}

		operations.OpenFile(ver.StudioPro)
	}

	return nil
}
