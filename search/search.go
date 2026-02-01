package search

import (
	"errors"
	"fmt"
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
