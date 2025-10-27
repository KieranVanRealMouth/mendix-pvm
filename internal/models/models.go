package models

import "fmt"

// Version represents a Mendix Studio Pro version
type Version struct {
	Major       int
	Minor       int
	Patch       int
	Build       int
	Path        string
	ModelerPath string
	MxPath      string
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", v.Major, v.Minor, v.Patch, v.Build)
}

func (v Version) ShortString() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Project represents a Mendix project
type Project struct {
	Name    string
	Path    string
	MprPath string
}

// Config represents the application configuration
type Config struct {
	VersionDirectory   string   `json:"version_directory"`
	ProjectDirectories []string `json:"project_directories"`
}

// SearchResult represents a search result that can be either a version or project
type SearchResult struct {
	Type    string // "version" or "project"
	Version *Version
	Project *Project
}

func (sr SearchResult) Display() string {
	if sr.Type == "version" {
		return fmt.Sprintf("[Version] %s", sr.Version.String())
	}
	return fmt.Sprintf("[Project] %s", sr.Project.Name)
}
