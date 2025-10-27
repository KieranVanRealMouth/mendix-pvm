package search

import (
	"strings"

	"mendix-project-manager/internal/models"
)

// Search searches for versions and projects matching the query
func Search(query string, versions []models.Version, projects []models.Project) []models.SearchResult {
	var results []models.SearchResult

	if query == "" {
		// Return all results
		for i := range versions {
			results = append(results, models.SearchResult{
				Type:    "version",
				Version: &versions[i],
			})
		}
		for i := range projects {
			results = append(results, models.SearchResult{
				Type:    "project",
				Project: &projects[i],
			})
		}
		return results
	}

	queryParts := strings.Fields(strings.ToLower(query))

	// Search versions
	for i := range versions {
		if matchesVersion(versions[i], queryParts) {
			results = append(results, models.SearchResult{
				Type:    "version",
				Version: &versions[i],
			})
		}
	}

	// Search projects
	for i := range projects {
		if matchesProject(projects[i], queryParts) {
			results = append(results, models.SearchResult{
				Type:    "project",
				Project: &projects[i],
			})
		}
	}

	return results
}

// SearchVersions searches only for versions
func SearchVersions(query string, versions []models.Version) []models.Version {
	if query == "" {
		return versions
	}

	queryParts := strings.Fields(strings.ToLower(query))
	var results []models.Version

	for _, version := range versions {
		if matchesVersion(version, queryParts) {
			results = append(results, version)
		}
	}

	return results
}

// SearchProjects searches only for projects
func SearchProjects(query string, projects []models.Project) []models.Project {
	if query == "" {
		return projects
	}

	queryParts := strings.Fields(strings.ToLower(query))
	var results []models.Project

	for _, project := range projects {
		if matchesProject(project, queryParts) {
			results = append(results, project)
		}
	}

	return results
}

func matchesVersion(version models.Version, queryParts []string) bool {
	versionStr := strings.ToLower(version.String())

	for _, part := range queryParts {
		if !strings.Contains(versionStr, part) {
			return false
		}
	}
	return true
}

func matchesProject(project models.Project, queryParts []string) bool {
	projectName := strings.ToLower(project.Name)
	// Normalize project name for better matching
	projectName = strings.ReplaceAll(projectName, "-", " ")
	projectName = strings.ReplaceAll(projectName, "_", " ")

	for _, part := range queryParts {
		if !strings.Contains(projectName, part) {
			return false
		}
	}
	return true
}
