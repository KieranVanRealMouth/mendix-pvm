package main

import (
	"fmt"
	"mendix-pvm/config"
	"mendix-pvm/convert"
	"mendix-pvm/project"
	"mendix-pvm/ui"
	"strings"
		Use:   "convert [additional search terms...]",
		Short: "Convert Mendix projects to a different Studio Pro version",
		Long: `Convert Mendix projects to a specified Studio Pro version.

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
		pathProjectOnly bool
		pathVersionOnly bool
	)

	var rootCmd = &cobra.Command{
		Use:   "mx",
		Short: "Manage Mendix Studio Pro versions and projects",
		Long: `mx is a command-line tool for managing Mendix Studio Pro installations and projects.

AVAILABLE COMMANDS:
  list      List and search for Mendix projects and Studio Pro versions
  open      Open a Mendix project or launch a Studio Pro version
  path      Print the full directory path of a project or Studio Pro version
  convert   Convert Mendix projects to a different Studio Pro version
  config    Open and edit the configuration file
  help      Display help information for any command

SEARCH BEHAVIOR:
All commands support case-insensitive fuzzy searching:
- Search terms match anywhere in project/version names and paths
- Multiple search terms must all match (AND logic)
- Version searches match Studio Pro version numbers (e.g., "10.6", "9.24")
- Project searches match against project folder names and paths

CONFIGURATION:
The tool searches in directories configured via 'mx config':
- Project directory: Where Mendix project folders are located
- Version directory: Where Studio Pro installations are stored

EXAMPLES:
  mx list                              # List all projects and versions
  mx list myapp                        # Search projects and versions containing "myapp"
  mx open 10.6                         # Open Studio Pro 10.6
  mx open MyProject                    # Open project "MyProject"
  mx path MyApp --project              # Print project directory path
  mx convert -p MyProject -v 10.10     # Convert project to version 10.10
  mx convert --project=MyApp --version=10.24  # Convert using equals syntax
  mx config                            # Edit configuration

For detailed help on any command, use: mx [command] --help
`,
	}

	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Open the configuration file",
		Long: `Opens the Mendix Project Manager configuration file in your system's default text editor.

The configuration file is a JSON file that controls:

PROJECT DIRECTORY:
  Path where the tool searches for Mendix project folders.
  Projects are identified by the presence of .mpr files.

VERSION DIRECTORY:
  Path where Mendix Studio Pro installations are located.
  Typically contains folders like "10.6.0.12345", "9.24.1.12345", etc.

AFTER EDITING:
- Save the file to apply changes
- Changes take effect immediately on next command
- Invalid paths will cause errors when running other commands

USAGE:
  mx config                        # Open config file for editing

NOTES:
- The config file location is platform-specific
- Ensure directories use proper path separators for your OS
- Both absolute and relative paths are supported`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Open(cfg); err != nil {
				return fmt.Errorf("An error occured while trying to open the config\n%w", err)
			}
			return nil
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list [search terms...]",
		Short: "List Mendix projects and Studio Pro versions",
		Long: `Search and display Mendix projects and/or installed Studio Pro versions.

SEARCH BEHAVIOR:
- Without arguments: Lists all available projects and versions
- With arguments: Filters results using case-insensitive substring matching
- Multiple search terms: All terms must match (AND logic)
- Matches against: Project/version names and full directory paths

OUTPUT FORMAT:
- Projects and versions are displayed in a formatted list
- Each item shows the full path
- Projects are identified by containing .mpr files
- Versions show the Studio Pro installation directory

AVAILABLE OPTIONS:
  -p, --project    Search and display only Mendix projects
                   (excludes Studio Pro versions from results)

  -v, --version    Search and display only Studio Pro versions
                   (excludes projects from results)

FLAG COMBINATIONS:
  (no flags)       Search both projects and versions (default)
  -p               Search only projects
  -v               Search only versions
  -p -v            Search both projects and versions (same as default)

EXAMPLES:
  mx list                          # Display all projects and versions
  mx list widget                   # Find projects/versions containing "widget"
  mx list 10.6                     # Find items containing "10.6"
  mx list 10.6 -v                  # Find only Studio Pro 10.6 versions
  mx list Customer -p              # Find only projects with "Customer" in path
  mx list myapp dev                # Find items containing both "myapp" AND "dev"
  mx list -p                       # Show all projects only
  mx list -v                       # Show all Studio Pro versions only

NOTES:
- Searches are case-insensitive
- Returns "No matches found." if no results match the criteria
- Results are not opened/executed, only displayed`,
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
		Use:   "open [search terms...]",
		Short: "Open a Mendix project or Studio Pro version",
		Long: `Open a Mendix project file or launch a Studio Pro version.

BASIC BEHAVIOR:
- Without arguments: Opens the first Studio Pro version found in configured directory
- With arguments: Searches for matching projects/versions using case-insensitive matching
- Single match: Opens immediately without confirmation
- Multiple matches: Displays list and requires refinement (unless --all is used)
- Maximum opens with --all: 3 items (safety limit to prevent opening too many)

OPENING BEHAVIOR:
- Projects: Launches the .mpr file using the associated Studio Pro version
- Versions: Starts the Studio Pro modeler.exe without opening a project
- Each item opens asynchronously (command returns immediately)

AVAILABLE OPTIONS:
  -p, --project    Search only Mendix projects
                   (excludes Studio Pro versions from search)

  -v, --version    Search only Studio Pro versions
                   (excludes projects from search)

  -a, --all        Open all matching results (maximum 3)
                   Use this when multiple matches are found
                   Opens items sequentially

FLAG COMBINATIONS:
  (no flags)       Search both projects and versions (default)
  -p               Search only projects
  -v               Search only versions
  -p -v            Search both (same as default)
  -a               Apply action to up to 3 matches
  -p -a            Open up to 3 matching projects
  -v -a            Open up to 3 matching versions

EXAMPLES:
  mx open                          # Open first Studio Pro version found
  mx open MyApp                    # Open project "MyApp" (if unique match)
  mx open 10.6                     # Open Studio Pro 10.6
  mx open 10.6 -v                  # Open Studio Pro 10.6 (versions only)
  mx open Customer -p              # Open project with "Customer" in path
  mx open MyApp --project          # Search only projects for "MyApp"
  mx open 10 -v -a                 # Open up to 3 Studio Pro versions matching "10"
  mx open widget                   # Search projects and versions for "widget"
  mx open myapp dev -p             # Find project matching both "myapp" AND "dev"

MULTIPLE MATCHES:
When multiple items match without --all flag:
- All matches are displayed in a list
- Command exits with instructions
- Options: Refine search terms or use --all/-a flag

NOTES:
- Search is case-insensitive
- Returns "No matches found." if no results match
- With --all, shows count when more than 3 matches exist
- Each opened item runs independently in a new process`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projects, versions, err := searchTargets(cfg, args, openProjectOnly, openVersionOnly)
			if err != nil {
				return err
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

	var pathCmd = &cobra.Command{
		Use:   "path <search terms...>",
		Short: "Print the directory path of a project or Studio Pro version",
		Long: `Resolve and print the full directory path of a single Mendix project or Studio Pro version.

PURPOSE:
Designed for scripting and command-line integration. Outputs the absolute directory path
to stdout, which can be used with cd, file operations, or other command-line tools.

BEHAVIOR:
- Searches for projects and/or versions using provided search terms
- Requires EXACTLY ONE match (strict requirement)
- Zero matches: Returns error "no matches found"
- Multiple matches: Displays list and returns error with match count
- Success: Prints the full absolute directory path to stdout

OUTPUT:
- Projects: Path to the project root directory (containing .mpr file)
- Versions: Path to the Studio Pro installation directory (containing modeler.exe)
- Format: Absolute path with platform-specific separators
- No trailing slash or newline (clean output for scripting)

AVAILABLE OPTIONS:
  -p, --project    Search only Mendix projects
                   (excludes Studio Pro versions from search)

  -v, --version    Search only Studio Pro versions
                   (excludes projects from search)

FLAG COMBINATIONS:
  (no flags)       Search both projects and versions (default)
  -p               Search only projects
  -v               Search only versions
  -p -v            Search both (same as default)

EXAMPLES:
  Basic usage:
    mx path MyApp                      # Print path to "MyApp" project
    mx path 10.6 -v                    # Print path to Studio Pro 10.6
    mx path Customer -p                # Print path to Customer project
    mx path 10.6 --version             # Print path to Studio Pro 10.6

  Shell integration (PowerShell):
    cd $(mx path MyApp)                # Change to project directory
    cd $(mx path 10.10 --version)      # Change to Studio Pro directory
    $projPath = mx path MyApp -p       # Store path in variable

  Shell integration (Bash/Unix):
    cd "$(mx path MyApp)"
    cd "$(mx path next dev --project)"

  Complex searches:
    mx path customer portal -p         # Find project matching both terms
    mx path 10.6.0 -v                  # Find specific version

ERROR HANDLING:
- No matches: Exits with error code and message
- Multiple matches: Shows list of all matches and exits with error
- Invalid path: Returns error if matched item doesn't exist

NOTES:
- Search is case-insensitive
- Requires at least one search term (path cannot list all items)
- Useful for automation, CI/CD pipelines, and shell scripting
- Output is designed to be machine-readable (no extra formatting)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projects, versions, err := searchTargets(cfg, args, pathProjectOnly, pathVersionOnly)
			if err != nil {
				return err
			}

			fullList := append(projects, versions...)
			if len(fullList) == 0 {
				return fmt.Errorf("no matches found")
			}
			if len(fullList) > 1 {
				sb := ui.List(fullList)
				cmd.Print(sb.String())
				return fmt.Errorf("multiple matches (%d). refine your arguments", len(fullList))
			}

			terminalPath, err := resolveTerminalPath(fullList[0])
			if err != nil {
				return err
			}
			fmt.Println(terminalPath)
			return nil
		},
	}

	var convertCmd = &cobra.Command{
		Use:   "convert --project <search> --version <version>",
		Short: "Convert Mendix projects to a different Studio Pro version",
		Long: `Convert one or more Mendix projects to a target Studio Pro version using Mendix's mx.exe tool.

REQUIRED FLAGS:
  -p, --project <search>     Project search term to find projects to convert
                             Searches project names and paths (case-insensitive)
                             Multiple words match with AND logic
                             Use quotes for values with spaces

  -v, --version <version>    Target Studio Pro version for conversion
                             Must match EXACTLY ONE installed Studio Pro version
                             Example: "10.10", "9.24", "10.6.0"

OPTIONAL FLAGS:
  -a, --all                  Convert ALL matching projects without confirmation
                             Default: Only converts first match if multiple found
                             No limit on number of conversions with this flag

FLAG SYNTAX:
Both space-separated and equals syntax are supported:
  --project MyApp            # Space-separated
  --project=MyApp            # Equals syntax
  --project="My App"         # Equals with quotes (recommended for spaces)
  -p MyApp                   # Short flag with space
  -p=MyApp                   # Short flag with equals

CONVERSION PROCESS:
1. Searches for projects matching --project argument
2. Validates that --version matches exactly one Studio Pro installation
3. If single project match or --all flag: proceeds with conversion
4. If multiple matches without --all: displays list and exits
5. For each project: Calls mx.exe from the target Studio Pro version
6. Displays progress and results for each conversion

EXIT CODES (per project):
  0 = Success: Project converted successfully
  1 = Internal error: mx.exe encountered an internal error
  2 = Option error: Invalid options passed to mx.exe
  3 = Conversion failed: Project conversion failed
 -1 = Execution error: Failed to start/execute mx.exe

BEHAVIOR DETAILS:
- Version must match exactly one installation (prevents ambiguity)
- Multiple version matches: Displays list and exits with error
- Projects are converted sequentially, not in parallel
- Each conversion shows: project name, progress (N/M), and result
- With --all: Continues converting remaining projects even if one fails
- Without --all: Stops on first error

EXAMPLES:
  Basic single project conversion:
    mx convert -p OrderApp -v 10.10
    mx convert --project CustomerPortal --version 10.6
    mx convert --project=MyApp --version=10.12.1

  Using equals syntax (recommended for clarity):
    mx convert --project="Order App" --version=10.10
    mx convert --project="Customer Portal" --version="10.6"
    mx convert -p=MyApp -v=10.24

  Projects with spaces in search term:
    mx convert --project="My Project" --version=10.24
    mx convert --project "My Project" --version 10.24

  Version with full number:
    mx convert --project=MyApp --version=10.24.1.12345

  Convert multiple matching projects:
    mx convert -p CustomerApp -v 10.10 -a
    mx convert --project=widget --version=9.24 --all
    mx convert --project="Customer App" --version=10.10 --all

  Convert with multiple search terms:
    mx convert -p "customer portal" -v 10.10    # Matches "customer" AND "portal"
    mx convert --project=myapp --version=10.6

IMPORTANT NOTES:
- BACKUP RECOMMENDED: Always backup projects before conversion
- Conversion is IRREVERSIBLE: Cannot downgrade using this tool
- Studio Pro must be closed: Close target version before converting
- Large projects: Conversion may take several minutes
- Version compatibility: Ensure target version supports your project features
- mx.exe location: Found in Studio Pro installation's modeler directory

TROUBLESHOOTING:
- "No version matches found": Install target Studio Pro version first
- "Multiple version matches": Refine --version to match exactly one
- "No project matches found": Check --project search term and config paths
- Exit code 3: Check Studio Pro logs for specific conversion errors
- Exit code 2: Verify mx.exe is compatible with the project type

SEQUENCE WITH --all FLAG:
  (1/3) Converting project: C:\Projects\App1
  Finished converting: C:\Projects\App1 (success)
  (2/3) Converting project: C:\Projects\App2
  Finished converting: C:\Projects\App2 (success)
  (3/3) Converting project: C:\Projects\App3
  Conversion failed: C:\Projects\App3`,
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

			// Combine --project flag and positional arguments for search
			var projectTokens []string
			if convertProject != "" {
				projectTokens = append(projectTokens, strings.Fields(convertProject)...)
			}
			if len(args) > 0 {
				projectTokens = append(projectTokens, args...)
			}
			projects, err := project.Search(cfg.ProjectDirectory, projectTokens)
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

	pathCmd.Flags().BoolVarP(&pathProjectOnly, "project", "p", false, "Limit search to projects only")
	pathCmd.Flags().BoolVarP(&pathVersionOnly, "version", "v", false, "Limit search to Studio Pro versions only")

	convertCmd.Flags().StringVarP(&convertVersion, "version", "v", "", "Mendix Studio Pro version to use for conversion")
	convertCmd.Flags().StringVarP(&convertProject, "project", "p", "", "Project filter (name or path segment) to convert")
	_ = convertCmd.MarkFlagRequired("version")
	_ = convertCmd.MarkFlagRequired("project")
	convertCmd.Flags().BoolVarP(&handleAll, "all", "a", false, "Convert all matching projects")

	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(pathCmd)
	rootCmd.AddCommand(convertCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func searchTargets(cfg *config.Config, args []string, projectOnly bool, versionOnly bool) ([]string, []string, error) {
	var (
		projects []string
		versions []string
		err      error
	)

	// If only --project is provided, search only projects.
	// If only --version is provided, search only versions.
	// If both or neither are provided, search both (default behavior).
	switch {
	case projectOnly && !versionOnly:
		projects, err = project.Search(cfg.ProjectDirectory, args)
		if err != nil {
			return nil, nil, fmt.Errorf("An error occured while trying to search project directory\n%w", err)
		}
		versions = []string{}

	case versionOnly && !projectOnly:
		versions, err = version.Search(cfg.VersionDirectory, args)
		if err != nil {
			return nil, nil, fmt.Errorf("An error occured while trying to search version directory\n%w", err)
		}
		projects = []string{}

	default:
		projects, err = project.Search(cfg.ProjectDirectory, args)
		if err != nil {
			return nil, nil, fmt.Errorf("An error occured while trying to search project directory\n%w", err)
		}

		versions, err = version.Search(cfg.VersionDirectory, args)
		if err != nil {
			return nil, nil, fmt.Errorf("An error occured while trying to search version directory\n%w", err)
		}
	}

	return projects, versions, nil
}

func resolveTerminalPath(path string) (string, error) {
	mprPath, err := project.FindMprAtRoot(path)
	if err != nil {
		return "", err
	}
	if mprPath != "" {
		return path, nil
	}

	if _, err := version.FindModelerSubdir(path); err != nil {
		return "", err
	}
	return path, nil
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
