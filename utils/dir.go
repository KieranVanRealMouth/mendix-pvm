package utils

import (
	"errors"
	"fmt"
	"os"
)

func IsDir(dirPath string) (bool, error) {
	info, err := os.Stat(dirPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, fmt.Errorf("path does not exist: %s", dirPath)
		}
		return false, fmt.Errorf("unable to stat path %s: %w", dirPath, err)
	}

	if info.IsDir() {
		return true, nil
	}
	return false, nil
}
