package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"mendix-project-manager/internal/models"
)

const configFileName = ".mendix-project-manager.json"

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configFileName), nil
}

// Load loads the configuration from the config file
func Load() (*models.Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config
			return &models.Config{
				VersionDirectory:   `C:\Program Files\Mendix\`,
				ProjectDirectories: []string{},
			}, nil
		}
		return nil, err
	}

	var config models.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Save saves the configuration to the config file
func Save(config *models.Config) error {
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
