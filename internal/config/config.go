package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	VersionDirectory string
	ProjectDirectory string
}

const configFileName = ".mendix-pvm.json"

func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configFileName), nil
}

func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			defaultVersionPath := `C:\Program Files\Mendix\`

			home, err := os.UserHomeDir()
			if err != nil {
				return nil, err
			}

			defaultProjectPath := filepath.Join(home, "Mendix")

			config := &Config{
				VersionDirectory: filepath.Clean(defaultVersionPath),
				ProjectDirectory: filepath.Clean(defaultProjectPath),
			}

			if err := config.Save(); err != nil {
				return nil, err
			}

			return config, nil
		}
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if err := config.Save(); err != nil {
		return nil, err
	}

	return &config, nil
}

func (config *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func (config *Config) Get(key string) (string, error) {
	switch key {
	case "versions":
		return config.VersionDirectory, nil
	case "projects":
		return config.ProjectDirectory, nil
	default:
		return "", fmt.Errorf("unknown key proveded: %s", key)
	}
}

func ValidateDirectoryPath(path string) error {
	// Clean the path to handle relative paths and resolve symlinks
	cleanPath := filepath.Clean(path)

	// Get absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", absPath)
		}
		return fmt.Errorf("cannot access path: %w", err)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absPath)
	}

	return nil
}

func (config *Config) Set(key, value string) error {
	// Validate that the value is a valid directory path
	if err := ValidateDirectoryPath(value); err != nil {
		return err
	}

	// Convert to absolute path for storage
	absPath, err := filepath.Abs(value)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	switch key {
	case "versions":
		config.VersionDirectory = absPath
	case "projects":
		config.ProjectDirectory = absPath
	default:
		return fmt.Errorf("unknown key: %s", key)
	}
	return nil
}
