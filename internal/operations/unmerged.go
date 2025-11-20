package operations

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetUnmergedBranches(path string) ([]string, error) {
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		fmt.Errorf("directory %s is not a git repository", path)
	}

	if err := gitFetchPrune(path); err != nil {
		fmt.Errorf("failed to fetch from origin for repository %s", path)
	}

	cmd := exec.Command("git", "branch", "-r", "--no-merged")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var branches []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip HEAD pointer
		if strings.Contains(line, "->") {
			continue
		}
		branches = append(branches, line)
	}

	return branches, nil
}

func gitFetchPrune(path string) error {
	cmd := exec.Command("git", "fetch", "origin", "--prune")
	cmd.Dir = path
	return cmd.Run()
}

func GetCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getCommitDiff(repoPath, currentRemoteBranch, remoteBranch string) (string, error) {
	compareSpec := fmt.Sprintf("%s...%s", currentRemoteBranch, remoteBranch)
	cmd := exec.Command("git", "log", compareSpec, "--oneline")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return "No commits", nil
	}
	return result, nil
}

func BuildOutput(repoPath, currentBranch string, unmergedBranches []string) string {
	var sb strings.Builder

	currentRemoteBranch := fmt.Sprintf("origin/%s", currentBranch)

	sb.WriteString(fmt.Sprintf("# Unmerged branches and commits of %s\n\n", currentBranch))
	sb.WriteString(fmt.Sprintf("Found %d unmerged branches.\n\n", len(unmergedBranches)))

	// Add list of unmerged branches
	if len(unmergedBranches) > 0 {
		sb.WriteString("## List of Unmerged Branches\n\n")
		for i, branch := range unmergedBranches {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, branch))
		}
		sb.WriteString("\n")
	}

	for _, branch := range unmergedBranches {
		sb.WriteString(fmt.Sprintf("## %s\n\n", branch))

		commits, err := getCommitDiff(repoPath, currentRemoteBranch, branch)
		if err != nil {
			sb.WriteString(fmt.Sprintf("Error getting commits: %v\n\n", err))
			continue
		}

		sb.WriteString(fmt.Sprintf("%s\n\n", commits))
	}

	return sb.String()
}
