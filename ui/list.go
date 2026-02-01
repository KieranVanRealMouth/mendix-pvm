package ui

import (
	"fmt"
	"path/filepath"
	"strings"
)

func List(items []string) strings.Builder {
	var sb strings.Builder
	for _, p := range items {
		name := filepath.Base(filepath.Clean(p)) // show only the dir name
		fmt.Fprintf(&sb, "- %s\n", name)
	}
	return sb
}
