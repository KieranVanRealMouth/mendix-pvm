package cmd

import (
	"fmt"
	"mendix-pvm/internal/operations"

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
	if projectFlag {
		if err := operations.FindAndOpenProject(args); err != nil {
			return err
		}
		return nil
	}

	if versionFlag {
		if err := operations.FindAndOpenVersion(args); err != nil {
			return err
		}
		return nil
	}

	if err := operations.FindAndOpenProject(args); err != nil {
		// if no project is found, try opening version
		fmt.Printf("\n")
		if err := operations.FindAndOpenVersion(args); err != nil {
			return err
		}
		return err
	}

	return nil
}
