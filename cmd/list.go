package cmd

import (
	"fmt"
	"mendix-pvm/internal/operations"

	"github.com/spf13/cobra"
)

var (
	listProjects bool
	listVersions bool
)

var listCmd = &cobra.Command{
	Use:   "list [query] [flags]",
	Short: "View projects and/or versions in the configured directories",
	Long:  "List all or a filtered list of projects/versions. Specify to only show versions or projects using the flags. The search can be done using a space seperated string.",
	Args:  cobra.RangeArgs(0, 10),
	RunE:  runList,
}

func init() {
	listCmd.Flags().BoolVar(&listProjects, "projects", false, "Only search for projects.")
	listCmd.Flags().BoolVar(&listVersions, "versions", false, "Only search for versions.")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	if listProjects {
		operations.ListProjects(args)
		return nil
	} else if listVersions {
		operations.ListVersions(args)
		return nil
	} else {
		operations.ListProjects(args)
		fmt.Printf("\n")
		operations.ListVersions(args)
		return nil
	}
}
