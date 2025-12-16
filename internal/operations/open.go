package operations

import (
	"fmt"
	"mendix-pvm/internal/project"
	"mendix-pvm/internal/version"
	"os/exec"
	"runtime"
)

func OpenFile(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	return nil
}

func FindAndOpenProject(args []string) error {
	projects, err := project.Search(args)
	if err != nil {
		return err
	}

	x := len(projects)
	switch {
	case x <= 0:
		return fmt.Errorf("no projects found for search query")
	case x == 1:
		OpenFile(projects[0].Mpr)
		fmt.Printf("Opened project: %s", projects[0].Name)
		return nil
	case x > 1:
		fmt.Printf("Found %v projects:\n", x)
		for _, proj := range projects {
			fmt.Printf("- %s\n", proj.Name)
		}
	default:
		return nil
	}

	return nil
}

func FindAndOpenVersion(args []string) error {
	versions, err := version.Search(args)
	if err != nil {
		return err
	}

	x := len(versions)
	switch {
	case x <= 0:
		return fmt.Errorf("no versions found for search query")
	case x == 1:
		OpenFile(versions[0].StudioPro)
		return nil
	case x > 1:
		fmt.Printf("Found %v versions:\n", x)
		for _, ver := range versions {
			fmt.Printf("- %s\n", ver.Name)
		}
	default:
		return nil
	}

	return nil
}
