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
