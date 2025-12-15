package operations

import (
	"fmt"
	"mendix-pvm/internal/project"
	"mendix-pvm/internal/version"
)

func ListProjects(args []string) {
	fmt.Printf("Projects:\n")
	projects, err := project.Search(args)
	if err != nil {
		fmt.Printf("an error occured searching projects: %v", err)
		return
	}

	for _, proj := range projects {
		fmt.Printf("- %s\n", proj.Name)
	}
}

func ListVersions(args []string) {
	versions, err := version.Search(args)
	if err != nil {
		fmt.Printf("an error occured searching versions: %v", err)
		return
	}

	fmt.Printf("Versions:\n")

	for _, ver := range versions {
		fmt.Printf("- %s\n", ver.Name)
	}
}
