package main

import (
	"fmt"
	"mendix-pvm/config"
	"mendix-pvm/convert"
	"mendix-pvm/platform"
	"mendix-pvm/project"
	"mendix-pvm/search"
	"mendix-pvm/ui"
	"mendix-pvm/version"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
		convertVersion   string
		convertProject   string
		handleAll        bool
		listProjectOnly  bool
		listVersionOnly  bool
		openProjectOnly  bool
		openVersionOnly  bool
		pathProjectOnly  bool
		pathVersionOnly  bool
		branchRepository string
		branchName       string
	)

	var rootCmd = &cobra.Command{
		Use:   "mx",
		Short: "Manage Mendix Studio Pro versions and projects",
		Long: `mx: Manage Mendix Studio Pro versions and projects.

Commands:
	list      List/search projects and Studio Pro versions
	open      Open a project or Studio Pro version
	path      Print directory path of a project/version
	convert   Convert projects to another Studio Pro version
	sync      Sync apps from the Mendix Platform
	config    Edit configuration file

Options:
	--project, -p   Limit to projects
	--version, -v   Limit to Studio Pro versions

Examples:
	mx list myapp
	mx open 10.6
	mx path MyApp --project
	mx convert -p MyApp -v 10.10
	mx sync
	mx config
	mx [command] --help (for details)
`,
	}

	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Open the configuration file",
		Long: `Edit the Mendix Project Manager config file (JSON).
Controls:
	- Project directory: Where projects are found
	- Version directory: Where Studio Pro installs are found

Usage:
	mx config

Edit and save to apply changes. Use valid paths.`,
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
		Long: `List Mendix projects and Studio Pro versions.

Options:
	--project, -p   Show projects only
	--version, -v   Show Studio Pro versions only

Examples:
	mx list
	mx list widget
	mx list 10.6 -v
	mx list Customer -p
`,
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
		Long: `Open a Mendix project or Studio Pro version.

Options:
	--project, -p   Search projects only
	--version, -v   Search Studio Pro versions only
	--all, -a       Open up to 3 matches

Examples:
	mx open MyApp
	mx open 10.6 -v
	mx open Customer -p
	mx open 10 -v -a
`,
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
		Long: `Print the directory path of a Mendix project or Studio Pro version.

Options:
	--project, -p   Search projects only
	--version, -v   Search Studio Pro versions only

Examples:
	mx path MyApp
	mx path 10.6 -v
	mx path Customer -p
	cd $(mx path MyApp)
`,
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
		Long: `Convert Mendix projects to a different Studio Pro version.

Required:
	--project, -p   Project to convert
	--version, -v   Target Studio Pro version
Optional:
	--all, -a       Convert all matches

Examples:
	mx convert -p MyApp -v 10.10
	mx convert --project="Order App" --version=10.10
	mx convert -p CustomerApp -v 10.10 -a

Notes:
	- Backup projects before converting
	- Conversion cannot be undone
`,
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

	var syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Sync apps from the Mendix Platform",
		Long: `Fetch all accessible Mendix apps from the Mendix Platform and
store their names and Git repository URLs in the local config.

Requires your Mendix User ID (OpenID) in the config and the MX_PAT environment
variable to be set.

Your User ID can be found in the Mendix Portal:
  1. Click your profile picture (top right)
  2. Go to User Settings
  3. Open the Personal Data tab
  4. Copy the value in the OpenID row

Run 'mx config' to set the User ID, or re-run the CLI setup to configure both.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			pat := os.Getenv("MX_PAT")
			if cfg.UserID == "" || pat == "" {
				return fmt.Errorf("Mendix credentials are not configured. Ensure your User ID is set in the config ('mx config') and MX_PAT is set as an environment variable.")
			}
			return platform.Sync(cmd.Context(), cfg, pat, func(s string) { cmd.Println(s) })
		},
	}

	var branchCmd = &cobra.Command{
		Use:   "branch",
		Short: "Manage Mendix app branches",
		Long: `Commands for working with Mendix application branches.

Commands:
    checkout    Clone a branch from a Mendix app repository

Examples:
    mx branch checkout -r "Approval Tool" -b feat/my-feature
`,
	}

	var checkoutCmd = &cobra.Command{
		Use:   "checkout",
		Short: "Clone a branch from a Mendix app's remote repository",
		Long: `Clone a single branch from a Mendix app's remote Git repository
into a new directory inside your configured projects directory.

The repository flag accepts a search query; the app name must contain the
provided string. If no matching app is found in the local config, a sync is
performed automatically before retrying.

Branch directories are created as:
    <App Name>-<branch_name>   (forward slashes in branch names become underscores)

Required:
    --repository, -r   Search query to identify the app
    --branch, -b       Branch name to clone

Examples:
    mx branch checkout -r "Approval Tool" -b feat/my-feature
    mx branch checkout --repository "Order" --branch main

Prerequisites (Studio Pro version control setup):
  Before using this command you must enable private version control in
  Mendix Studio Pro:
    1. Open Edit > Preferences > Version Control > Git.
    2. Enable "Enable private version control with Git".
    3. Enter a Name and Email for your Git identity.
    4. Enable "Use Windows credentials" so that the PAT token used during
       the initial clone is reused automatically for subsequent operations.

Troubleshooting:
  If you get an "access denied" or authentication error, automated PAT
  token detection may have failed. To recover:
    1. In Studio Pro, sign out (Edit > Sign Out).
    2. Re-run this command.
    3. Git will prompt for your username and password — enter your Mendix
       account username and a valid PAT token as the password.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			pat := os.Getenv("MX_PAT")

			// 1. Search current config
			matches := search.SearchApps(cfg.Apps, branchRepository)

			// 2. If no match, sync and retry
			if len(matches) == 0 {
				if cfg.UserID == "" || pat == "" {
					return fmt.Errorf("repository %q not found in config and Mendix credentials are not set; cannot sync", branchRepository)
				}
				cmd.Printf("Repository %q not found in config. Syncing...\n", branchRepository)
				if err := platform.Sync(cmd.Context(), cfg, pat, func(s string) { cmd.Println(s) }); err != nil {
					return fmt.Errorf("sync failed: %w", err)
				}
				matches = search.SearchApps(cfg.Apps, branchRepository)
			}

			// 3. Exit conditions
			if len(matches) == 0 {
				return fmt.Errorf("no repository found matching %q", branchRepository)
			}
			if len(matches) > 1 {
				var names []string
				for _, m := range matches {
					names = append(names, m.Name)
				}
				return fmt.Errorf("multiple repositories match %q: %s\nRefine your --repository query", branchRepository, strings.Join(names, ", "))
			}

			app := matches[0]

			// 4. Build destination directory path
			safeBranch := strings.ReplaceAll(branchName, "/", "_")
			dirName := app.Name + "-" + safeBranch
			destDir := filepath.Join(cfg.ProjectDirectory, dirName)

			// 5. Clone
			cmd.Printf("Cloning branch %q of %q into:\n  %s\n", branchName, app.Name, destDir)
			gitCmd := exec.CommandContext(
				cmd.Context(),
				"git", "clone",
				"--branch", branchName,
				"--single-branch",
				app.RepositoryURL,
				destDir,
			)
			gitCmd.Stdout = cmd.OutOrStdout()
			gitCmd.Stderr = cmd.ErrOrStderr()
			if err := gitCmd.Run(); err != nil {
				return fmt.Errorf("git clone failed: %w", err)
			}

			fetchNotesCmd := exec.CommandContext(
				cmd.Context(),
				"git", "-C", destDir,
				"fetch", "origin", "refs/notes/mx_metadata:refs/notes/mx_metadata",
			)
			fetchNotesCmd.Stdout = cmd.OutOrStdout()
			fetchNotesCmd.Stderr = cmd.ErrOrStderr()
			if err := fetchNotesCmd.Run(); err != nil {
				cmd.Printf("Warning: could not fetch Mendix metadata notes: %v\n", err)
			}

			cmd.Printf("Done. Branch available at: %s\n", destDir)
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
	rootCmd.AddCommand(syncCmd)

	checkoutCmd.Flags().StringVarP(&branchRepository, "repository", "r", "", "Search query to identify the app repository")
	checkoutCmd.Flags().StringVarP(&branchName, "branch", "b", "", "Branch name to clone")
	_ = checkoutCmd.MarkFlagRequired("repository")
	_ = checkoutCmd.MarkFlagRequired("branch")

	branchCmd.AddCommand(checkoutCmd)
	rootCmd.AddCommand(branchCmd)

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
