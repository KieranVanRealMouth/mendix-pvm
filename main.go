package main

import (
	"fmt"
	"mendix-pvm/config"
	"mendix-pvm/convert"
	"mendix-pvm/project"
	"mendix-pvm/ui"

	"mendix-pvm/version"
	"os"

	"github.com/spf13/cobra"
)

// TODO: Implement flags and shorthand flags for --project and --version
// - For list, this will allow the user to show or the projects only or the versions only
// - For open, this will only search projects or versions depending on the specified flag
// - If both are provided, the behavior is as the default behavior is
// - For convert it's already implemented in it's own specific way
// - For config, it's not needed
func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("An error occured while trying to load the config\n%w", err)
	}

	// initialize flags
	var (
		convertVersion  string
		convertProject  string
		handleAll       bool
		listProjectOnly bool
		listVersionOnly bool
		openProjectOnly bool
		openVersionOnly bool
	)

	var rootCmd = &cobra.Command{
		Use:   "mx",
		Short: "Mendix Project Manager",
		Long:  `A CLI tool for managing Mendix Studio Pro versions and projects.`,
	}

	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Open the config file.",
		Long:  "Open the config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Open(cfg); err != nil {
				return fmt.Errorf("An error occured while trying to open the config\n%w", err)
			}
			return nil
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "Search using arguments",
		Long:  "Provide any number of arguments to search Mendix projects and versions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				projects []string
				versions []string
				err      error
			)

			// If only --project is provided, search only projects.
			// If only --version is provided, search only versions.
			// If both or neither are provided, search both (default behavior).
			switch {
			case listProjectOnly && !listVersionOnly:
				projects, err = project.Search(cfg.ProjectDirectory, args)
				if err != nil {
					return fmt.Errorf("An error occured while trying to search project directory\n%w", err)
				}
				versions = []string{}

			case listVersionOnly && !listProjectOnly:
				versions, err = version.Search(cfg.VersionDirectory, args)
				if err != nil {
					return fmt.Errorf("An error occured while trying to search version directory\n%w", err)
				}
				projects = []string{}

			default:
				projects, err = project.Search(cfg.ProjectDirectory, args)
				if err != nil {
					return fmt.Errorf("An error occured while trying to search project directory\n%w", err)
				}

				versions, err = version.Search(cfg.VersionDirectory, args)
				if err != nil {
					return fmt.Errorf("An error occured while trying to search version directory\n%w", err)
				}
			}

			fullList := append(projects, versions...)

			if len(fullList) == 0 {
				cmd.Println("No matches found.")
				return nil
			}

			sb := ui.List(fullList)
			cmd.Print(sb.String())
			return nil
		},
	}

	var openCmd = &cobra.Command{
		Use:   "open",
		Short: "Open Mendix",
		Long:  "Open any Mendix project or version by providing arguments (seperated by spaces). When no arguments are provided, the latest version of studio pro is opened. If arguments are provided, only 1 project can be opened at a time, meaning the arguments must be improved to match only 1 result. If the --all flag or -a shorthand is provided, all matching results are opened. A maximum of 3 items can be opened at a time.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				projects []string
				versions []string
				err      error
			)

			// If only --project is provided, search only projects.
			// If only --version is provided, search only versions.
			// If both or neither are provided, search both (default behavior).
			switch {
			case openProjectOnly && !openVersionOnly:
				projects, err = project.Search(cfg.ProjectDirectory, args)
				if err != nil {
					return fmt.Errorf("An error occured while trying to search project directory\n%w", err)
				}
				versions = []string{}

			case openVersionOnly && !openProjectOnly:
				versions, err = version.Search(cfg.VersionDirectory, args)
				if err != nil {
					return fmt.Errorf("An error occured while trying to search version directory\n%w", err)
				}
				projects = []string{}

			default:
				projects, err = project.Search(cfg.ProjectDirectory, args)
				if err != nil {
					return fmt.Errorf("An error occured while trying to search project directory\n%w", err)
				}

				versions, err = version.Search(cfg.VersionDirectory, args)
				if err != nil {
					return fmt.Errorf("An error occured while trying to search version directory\n%w", err)
				}
			}

			fullList := append(projects, versions...)

			if len(fullList) == 0 {
				cmd.Println("No matches found.")
				return nil
			}

			if len(fullList) == 1 && !handleAll {
				OpenProjectOrVersion(fullList[0], cmd)
				return nil
			}

			if !handleAll {
				sb := ui.List(fullList)
				cmd.Print(sb.String())
				cmd.Printf("\nMultiple matches (%d). Refine your arguments or pass --all/-a to open the first 3 match(es).\n", len(fullList))
				return nil
			}

			maxOpen := 3
			toOpen := len(fullList)

			if toOpen > maxOpen {
				cmd.Printf("Opening first %d matches (of %d).\n", maxOpen, len(fullList))
			}

			if toOpen > maxOpen {
				toOpen = maxOpen
			}

			for i := 0; i < toOpen; i++ {
				item := fullList[i]
				OpenProjectOrVersion(item, cmd)
			}

			return nil
		},
	}

	var convertCmd = &cobra.Command{
		Use:   "convert",
		Short: "Convert Mendix projects from 1 version to another",
		Long:  "By default only convert 1 Mendix project to the chosen version. Use the --all flag or -a shorthand to convert all matches to the query.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if convertVersion == "" || convertProject == "" {
				return fmt.Errorf("--version and --project are required")
			}

			versions, err := version.Search(cfg.VersionDirectory, []string{convertVersion})
			if err != nil {
				return fmt.Errorf("An error occured while trying to search version directory\n%w", err)
			}
			if len(versions) == 0 {
				return fmt.Errorf("No version matches found for %q", convertVersion)
			}
			// must be exactly 1 match as there's only 1 mx.exe tool to call to convert to
			if len(versions) > 1 {
				sb := ui.List(versions)
				cmd.Print(sb.String())
				fmt.Printf("Multiple version matches (%d) for %q. Please refine --version.", len(versions), convertVersion)
				return nil
			}

			// may be 1 or many depending on --all
			projects, err := project.Search(cfg.ProjectDirectory, []string{convertProject})
			if err != nil {
				return fmt.Errorf("An error occured while trying to search project directory\n%w", err)
			}
			if len(projects) == 0 {
				cmd.Println("No project matches found.")
				return nil
			}

			if !handleAll && len(projects) > 1 {
				sb := ui.List(projects)
				cmd.Print(sb.String())
				cmd.Printf("\nMultiple matches (%d). Refine your --project argument or pass --all/-a to convert all matches.\n", len(projects))
				return nil
			}

			toConvert := projects
			if !handleAll {
				toConvert = projects[:1]
			}

			vPath := versions[0]
			for i, p := range toConvert {
				cmd.Printf("(%d/%d) Converting project: %s\n", i+1, len(toConvert), p)

				exitCode, err := convert.Convert(vPath, p)
				switch exitCode {
				case 0:
					cmd.Printf("Finished converting: %s (success)\n", p)
				case 1:
					cmd.Printf("Internal error during convertsion: %s \n", p)
				case 2:
					cmd.Printf("Error with options: %s\n", p)
				case 3:
					cmd.Printf("Conversion failed: %s\n", p)
				default:
					// -1 means start/exec error in our wrapper; any other code is unexpected but log it.
					return fmt.Errorf("Failed to convert: %s (unexpected exit code %d)\n", p, exitCode)
				}

				if err != nil && !handleAll {
					return err
				}
			}

			return nil
		},
	}

	// COMMAND CALLS
	listCmd.Flags().BoolVarP(&listProjectOnly, "project", "p", false, "Show projects only")
	listCmd.Flags().BoolVarP(&listVersionOnly, "version", "v", false, "Show Studio Pro versions only")

	openCmd.Flags().BoolVarP(&openProjectOnly, "project", "p", false, "Limit search to projects only")
	openCmd.Flags().BoolVarP(&openVersionOnly, "version", "v", false, "Limit search to Studio Pro versions only")
	openCmd.Flags().BoolVarP(
		&handleAll,
		"all",
		"a",
		false,
		"Perform action for all matching results (might be limited depending on the load a command creates)",
	)

	convertCmd.Flags().StringVarP(&convertVersion, "version", "v", "", "Mendix Studio Pro version to use for conversion")
	convertCmd.Flags().StringVarP(&convertProject, "project", "p", "", "Project filter (name or path segment) to convert")
	_ = convertCmd.MarkFlagRequired("version")
	_ = convertCmd.MarkFlagRequired("project")
	convertCmd.Flags().BoolVarP(&handleAll, "all", "a", false, "Convert all matching projects")

	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(convertCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func OpenProjectOrVersion(path string, cmd *cobra.Command) error {
	ok, err := project.FindMprAtRoot(path)
	if err != nil {
		return err
	}

	if ok != "" {
		if err := project.Open(path); err != nil {
			return fmt.Errorf("failed to open project %q: %w", path, err)
		}

		cmd.Printf("Opening project: %s\n", path)
		return nil
	} else {
		if err := version.Open(path); err != nil {
			return fmt.Errorf("failed to open Studio Pro at %q: %w", path, err)
		}

		cmd.Printf("Opening Studio Pro: %s\n", path)

		return nil
	}
}
