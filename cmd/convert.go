package cmd

import (
	"fmt"
	"mendix-pvm/internal/config"
	"mendix-pvm/internal/operations"
	"mendix-pvm/internal/project"
	"mendix-pvm/internal/search"
	"mendix-pvm/internal/version"
	"sync"

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

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	semaphore := make(chan struct{}, 5)

	for _, result := range results {
		wg.Add(1)

		semaphore <- struct{}{}

		go func(resultPath string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			proj, err := project.CreateProject(result.Path)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("failed to create project %s: %w", result.Path, err))
				mu.Unlock()
				return
			}

			if err := operations.ConvertProject(ver, proj); err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("failed to convert project %s: %w", proj.Name, err))
				mu.Unlock()
				return
			}
		}(result.Path)
	}

	wg.Wait()

	if len(errors) > 0 {
		fmt.Printf("An error occured while attempting to convert projects:\n")
		fmt.Printf("- Success: %d", len(results)-len(errors))
		fmt.Printf("- Failed: %d", len(errors))
		fmt.Printf("- Total: %d", len(results))
		return fmt.Errorf("conversion completed with %d errors", len(errors))
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
