package search

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SearchResult struct {
	Path string
	Name string
}

func Search(directory string, query []string) ([]SearchResult, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var results []SearchResult

	numberOfQueries := len(query)
	if numberOfQueries > 0 {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			name := entry.Name()
			normalizedName := strings.ToLower(name)

			var numberOfMatches int

			for _, item := range query {
				normalizedQuery := strings.ToLower(item)

				if strings.Contains(normalizedName, normalizedQuery) {
					numberOfMatches += 1
				}
			}

			if numberOfMatches == numberOfQueries {
				result, err := createSearchResult(filepath.Join(directory, name))
				if err != nil {
					return nil, err
				}

				results = append(results, result)
			}
		}
	} else {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			result, err := createSearchResult(filepath.Join(directory, entry.Name()))
			if err != nil {
				return nil, err
			}

			results = append(results, result)
		}
	}
	return results, nil
}

func createSearchResult(path string) (SearchResult, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return SearchResult{}, fmt.Errorf("not able to create absolute path: %w", err)
	}

	return SearchResult{Path: absolutePath, Name: filepath.Base(path)}, nil
}
