package cmd

import (
	"fmt"
	"mendix-pvm/internal/config"
	"mendix-pvm/internal/search"

	"github.com/spf13/cobra"
)

var (
	projects bool
	versions bool
)

var listCmd = &cobra.Command{
	Use:   "list [query] [flags]",
	Short: "View projects and/or versions in the configured directories",
	Long:  "List all or a filtered list of projects/versions. Specify to only show versions or projects using the flags. The search can be done using a space seperated string.",
	Args:  cobra.RangeArgs(0, 10),
	RunE:  runList,
}

func init() {
	listCmd.Flags().BoolVar(&projects, "projects", false, "Only search for projects.")
	listCmd.Flags().BoolVar(&versions, "versions", false, "Only search for versions.")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if projects { // log only projects
		if err := searchAndLogDirectory(cfg.ProjectDirectory, args); err != nil {
			return err
		}
		return nil
	} else if versions { // log only version
		if err := searchAndLogDirectory(cfg.VersionDirectory, args); err != nil {
			return err
		}

		return nil
	} else { // log both versions and projects
		if err := searchAndLogDirectory(cfg.VersionDirectory, args); err != nil {
			return err
		}

		if err := searchAndLogDirectory(cfg.ProjectDirectory, args); err != nil {
			return err
		}

		return nil
	}
}

func searchAndLogDirectory(directory string, args []string) error {
	results, err := search.Search(directory, args)
	if err != nil {
		return fmt.Errorf("failed to search directory %s: %w", directory, err)
	}

	for _, result := range results {
		fmt.Println("- " + result.Name)
	}

	return nil
}
