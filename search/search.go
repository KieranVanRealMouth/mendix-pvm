package search

import (
	"errors"
	"fmt"
	"mendix-pvm/config"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func SearchDir(searchPath string, query []string) ([]string, error) {
	info, err := os.Stat(searchPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("path does not exist: %s", searchPath)
		}
		return nil, fmt.Errorf("unable to stat path %s: %w", searchPath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", searchPath)
	}

	var tokens []string
	for _, q := range query {
		n := normalize(q)
		if n != "" {
			tokens = append(tokens, n)
		}
	}

	entries, err := os.ReadDir(searchPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read directory %s: %w", searchPath, err)
	}

	var matches []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		dirName := e.Name()
		normName := normalize(dirName)
		if matchAllTokens(normName, tokens) {
			// Print or collect absolute path? Keeping name and relative path for clarity:
			matches = append(matches, filepath.Join(searchPath, dirName))
		}
	}

	return matches, nil
}

func normalize(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func matchAllTokens(normName string, tokens []string) bool {
	if len(tokens) == 0 {
		return true
	}
	for _, t := range tokens {
		if !strings.Contains(normName, t) {
			return false
		}
	}
	return true
}

// FilterPaths filters an already-loaded list of directory paths by base name,
// using the same normalization and token matching as SearchDir.
func FilterPaths(paths []string, query string) []string {
	var tokens []string
	for _, q := range strings.Fields(query) {
		if n := normalize(q); n != "" {
			tokens = append(tokens, n)
		}
	}
	if len(tokens) == 0 {
		return paths
	}
	var matches []string
	for _, p := range paths {
		if matchAllTokens(normalize(filepath.Base(p)), tokens) {
			matches = append(matches, p)
		}
	}
	return matches
}

// FilterStrings filters a list of plain strings by their full value,
// using the same normalization and token matching as FilterPaths but without
// stripping path separators. Use this for branch names that contain slashes.
func FilterStrings(items []string, query string) []string {
	var tokens []string
	for _, q := range strings.Fields(query) {
		if n := normalize(q); n != "" {
			tokens = append(tokens, n)
		}
	}
	if len(tokens) == 0 {
		return items
	}
	var matches []string
	for _, s := range items {
		if matchAllTokens(normalize(s), tokens) {
			matches = append(matches, s)
		}
	}
	return matches
}

// SearchApps returns every app whose name matches all tokens derived from query.
func SearchApps(apps []config.App, query string) []config.App {
	tokens := []string{}
	for _, q := range strings.Fields(query) {
		if n := normalize(q); n != "" {
			tokens = append(tokens, n)
		}
	}

	var matches []config.App
	for _, app := range apps {
		if matchAllTokens(normalize(app.Name), tokens) {
			matches = append(matches, app)
		}
	}
	return matches
}
