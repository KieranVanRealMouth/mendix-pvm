package branch

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"mendix-pvm/config"
)

func setProjectId(destDir string) error {
	// get config file at .git/config
	configPath := filepath.Join(destDir, ".git", "config")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read git config: %w", err)
	}
	lines := strings.Split(string(configData), "\n")

	// get remote url (looks for a line containing "url =")
	remoteUrl := ""
	for _, line := range lines {
		if strings.Contains(line, "url =") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				remoteUrl = strings.TrimSpace(parts[1])
				break
			}
		}
	}
	if remoteUrl == "" {
		return fmt.Errorf("remote url not found in git config")
	}

	// only run project-id logic for Mendix hosted repos
	if !strings.HasPrefix(remoteUrl, "https://git.api.mendix.com") {
		return nil
	}

	// normalize and extract last path segment as project id
	remoteUrl = strings.TrimSuffix(remoteUrl, ".git/")
	remoteUrl = strings.TrimSuffix(remoteUrl, ".git")
	// replace ':' with '/' to handle git@host:owner/repo.git
	remoteUrl = strings.ReplaceAll(remoteUrl, ":", "/")
	parts := strings.Split(remoteUrl, "/")
	projectId := parts[len(parts)-1]

	// look for existing sprintr-project-id entry
	pidIndex := -1
	for i, line := range lines {
		if strings.Contains(line, "sprintr-project-id") {
			pidIndex = i
			break
		}
	}

	if pidIndex != -1 {
		line := lines[pidIndex]
		if idx := strings.Index(line, "="); idx != -1 {
			existing := strings.TrimSpace(line[idx+1:])
			if existing == projectId {
				// already set correctly
				return nil
			}
		}
		// update existing line (either had no value or different value)
		lines[pidIndex] = fmt.Sprintf("\tsprintr-project-id = %s", projectId)
		newConfigData := strings.Join(lines, "\n")
		if err := os.WriteFile(configPath, []byte(newConfigData), 0644); err != nil {
			return fmt.Errorf("failed to write git config: %w", err)
		}
		return nil
	}

	// not found — add it under [mendix] section (create section if needed)
	mendixSectionIndex := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "[mendix]" {
			mendixSectionIndex = i
			break
		}
	}
	projectIdLine := fmt.Sprintf("\tsprintr-project-id = %s", projectId)
	if mendixSectionIndex == -1 {
		// ensure file ends with a newline before appending
		if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
			lines = append(lines, "")
		}
		lines = append(lines, "[mendix]", projectIdLine)
	} else {
		// insert after the [mendix] header
		insertAt := mendixSectionIndex + 1
		lines = append(lines[:insertAt], append([]string{projectIdLine}, lines[insertAt:]...)...)
	}

	newConfigData := strings.Join(lines, "\n")
	if err := os.WriteFile(configPath, []byte(newConfigData), 0644); err != nil {
		return fmt.Errorf("failed to write git config: %w", err)
	}

	return nil
}

// Checkout clones a single branch from the app's remote repository into destDir
// and fetches Mendix metadata notes.
func Checkout(ctx context.Context, app config.App, branchName, destDir string, stdout, stderr io.Writer) error {
	gitCmd := exec.CommandContext(
		ctx,
		"git", "clone",
		"--branch", branchName,
		app.RepositoryURL,
		destDir,
	)
	gitCmd.Stdout = stdout
	gitCmd.Stderr = stderr
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	// attempt to set sprintr-project-id in .git/config for Mendix hosted repos
	if err := setProjectId(destDir); err != nil {
		fmt.Fprintf(stdout, "Warning: could not set sprintr-project-id: %v\n", err)
	}

	fetchNotesCmd := exec.CommandContext(
		ctx,
		"git", "-C", destDir,
		"fetch", "origin", "refs/notes/mx_metadata:refs/notes/mx_metadata",
	)
	fetchNotesCmd.Stdout = stdout
	fetchNotesCmd.Stderr = stderr
	if err := fetchNotesCmd.Run(); err != nil {
		fmt.Fprintf(stdout, "Warning: could not fetch Mendix metadata notes: %v\n", err)
	}

	return nil
}

// Create creates or reuses a branch in the app's remote repository.
// If the branch already exists on the remote, it is cloned. Otherwise, it is
// created from baseBranch on the remote and then checked out.
func Create(ctx context.Context, cfg *config.Config, app config.App, branchName, baseBranch string, stdout, stderr io.Writer) error {
	safeBranch := strings.ReplaceAll(branchName, "/", "_")
	dirName := app.Name + "-" + safeBranch
	destDir := filepath.Join(cfg.ProjectDirectory, dirName)

	// Check if destination directory already exists
	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("branch directory already exists: %s\nThe branch appears to be already checked out.", destDir)
	}

	// Check whether the branch exists on the remote
	out, err := exec.CommandContext(ctx, "git", "ls-remote", "--heads", app.RepositoryURL, branchName).Output()
	branchExists := err == nil && len(strings.TrimSpace(string(out))) > 0

	if branchExists {
		fmt.Fprintf(stdout, "Branch %q already exists on remote. Cloning into:\n  %s\n", branchName, destDir)
		if err := Checkout(ctx, app, branchName, destDir, stdout, stderr); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Done. Branch available at: %s\n", destDir)
		return nil
	}

	// Branch does not exist — create it on the remote, then check it out
	fmt.Fprintf(stdout, "Branch %q does not exist on remote. Creating from %q into:\n  %s\n", branchName, baseBranch, destDir)

	run := func(dir string, args ...string) error {
		c := exec.CommandContext(ctx, args[0], args[1:]...)
		c.Dir = dir
		c.Stdout = stdout
		c.Stderr = stderr
		return c.Run()
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := run(destDir, "git", "init"); err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}

	if err := run(destDir, "git", "remote", "add", "origin", app.RepositoryURL); err != nil {
		return fmt.Errorf("git remote add failed: %w", err)
	}

	if err := run(destDir, "git", "fetch", "origin", baseBranch); err != nil {
		return fmt.Errorf("git fetch base branch %q failed: %w", baseBranch, err)
	}

	if err := run(destDir, "git", "push", "origin", "FETCH_HEAD:refs/heads/"+branchName); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	if err := run(destDir, "git", "fetch", "origin", branchName); err != nil {
		return fmt.Errorf("git fetch new branch %q failed: %w", branchName, err)
	}

	if err := run(destDir, "git", "checkout", "-b", branchName, "--track", "origin/"+branchName); err != nil {
		return fmt.Errorf("git checkout failed: %w", err)
	}

	// attempt to set sprintr-project-id in .git/config for Mendix hosted repos
	if err := setProjectId(destDir); err != nil {
		fmt.Fprintf(stdout, "Warning: could not set sprintr-project-id: %v\n", err)
	}

	// Fetch Mendix metadata notes
	fetchNotesCmd := exec.CommandContext(
		ctx,
		"git", "-C", destDir,
		"fetch", "origin", "refs/notes/mx_metadata:refs/notes/mx_metadata",
	)
	fetchNotesCmd.Stdout = stdout
	fetchNotesCmd.Stderr = stderr
	if err := fetchNotesCmd.Run(); err != nil {
		fmt.Fprintf(stdout, "Warning: could not fetch Mendix metadata notes: %v\n", err)
	}

	fmt.Fprintf(stdout, "Done. Branch available at: %s\n", destDir)
	return nil
}
