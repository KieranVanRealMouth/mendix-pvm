package cmd

import (
	"fmt"
	"mendix-pvm/internal/config"
	"mendix-pvm/internal/project"
	"mendix-pvm/internal/search"
	"mendix-pvm/internal/version"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var convertVersion string

var convertCmd = &cobra.Command{
	Use:   "convert [query]",
	Short: "Convert the projects matching the query to the specified version",
	Long:  "Converts all matched projects their local mpr file to the specified version without committing",
	RunE:  runConvert,
}

func init() {
	convertCmd.Flags().StringVar(&convertVersion, "version", "", "Version to convert to")

	if err := convertCmd.MarkFlagRequired("version"); err != nil {
		fmt.Printf("the version flag is required")
		return
	}

	rootCmd.AddCommand(convertCmd)
}

func runConvert(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ver, err := getVersion(cfg.VersionDirectory, convertVersion)
	if err != nil {
		return err
	}

	results, err := search.Search(cfg.ProjectDirectory, args)
	if err != nil {
		return err
	}

	for _, result := range results {
		proj, err := project.CreateProject(result.Path)
		if err != nil {
			return err
		}

		if err := convertProject(ver, proj); err != nil {
			return err
		}

	}

	return nil
}

func getVersion(directory string, arg string) (version.Version, error) {
	arguments := [1]string{arg}

	results, err := search.Search(directory, arguments[:])
	if err != nil {
		return version.Version{}, err
	}

	x := len(results)

	if x == 0 {
		return version.Version{}, fmt.Errorf("no results found")
	}

	if x > 1 {
		fmt.Printf("Found %v versions.\n", x)
		for _, res := range results {
			fmt.Printf("- %s\n", res.Name)
		}
		return version.Version{}, fmt.Errorf("please select a unique version")
	}

	return version.CreateVersion(results[0].Path)
}

func getExecutable(ver version.Version) (string, error) {
	executablePath := filepath.Join(ver.Directory, "modeler", "mx.exe")
	if _, err := os.Stat(executablePath); err != nil {
		return "", err
	}

	return executablePath, nil
}

func convertProject(ver version.Version, proj project.Project) error {
	fmt.Printf("Converting %s to version %s\n", proj.Name, ver.Name)
	mxPath, err := getExecutable(ver)
	if err != nil {
		return err
	}

	cmd := exec.Command(mxPath, "convert", "--in-place", proj.Directory)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("conversion failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("Finished converting\n")

	return nil
}
