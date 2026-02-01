package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const configFileName = ".mendix-pvm.json"

type Config struct {
	VersionDirectory string
	ProjectDirectory string
}

func create() (Config, error) {
	var versionDir string
	switch runtime.GOOS {
	case "windows":
		programFiles := os.Getenv("ProgramFiles")
		if programFiles != "" {
			versionDir = filepath.Join(programFiles, "Mendix")
		}
	case "darwin":
		return Config{}, fmt.Errorf("macOS is not supported")
	default:
		versionDir = ""
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("failed to determine user home directory: %w", err)
	}
	projectDir := filepath.Join(home, "Mendix")

	return Config{
		VersionDirectory: versionDir,
		ProjectDirectory: projectDir,
	}, nil
}

func (config *Config) save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func dirExists(path string) (ok bool, why string) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, "does not exist"
		}
		return false, fmt.Sprintf("cannot be accessed: %v", err)
	}
	if !info.IsDir() {
		return false, "is not a directory"
	}
	return true, ""
}

func validate(config Config) error {
	var issues []string

	if strings.TrimSpace(config.VersionDirectory) == "" {
		issues = append(issues, "versionDirectory is not set")
	} else {
		if ok, why := dirExists(config.VersionDirectory); !ok {
			issues = append(issues, fmt.Sprintf("versionDirectory %q %s", config.VersionDirectory, why))
		}
	}

	if strings.TrimSpace(config.ProjectDirectory) == "" {
		issues = append(issues, "projectDirectory is not set")
	} else {
		if ok, why := dirExists(config.ProjectDirectory); !ok {
			issues = append(issues, fmt.Sprintf("projectDirectory %q %s", config.ProjectDirectory, why))
		}
	}

	if len(issues) > 0 {
		return fmt.Errorf("invalid configuration: %s", strings.Join(issues, "; "))
	}

	return nil
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, configFileName), nil

}

func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	var config Config

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			config, err = create()
			if err != nil {
				return nil, err
			}

			if err := config.save(); err != nil {
				return nil, err
			}

			return &config, nil
		}
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if err := validate(config); err != nil {
		return nil, err
	}

	return &config, nil
}

func Open(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	switch runtime.GOOS {
	case "windows":
		return exec.Command("cmd", "/c", "start", "", configPath).Start()
	case "darwin":
		return exec.Command("open", configPath).Start()
	default:
		return exec.Command("xdg-open", configPath).Start()
	}
}
