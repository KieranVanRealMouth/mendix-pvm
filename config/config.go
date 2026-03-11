package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/term"
)

const configFileName = ".mendix-pvm.json"

type App struct {
	Name          string `json:"name"`
	RepositoryURL string `json:"repositoryUrl"`
}

type Config struct {
	VersionDirectory string `json:"versionDirectory"`
	ProjectDirectory string `json:"projectDirectory"`
	UserID           string `json:"userID"`
	Apps             []App  `json:"apps"`
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

	cfg := Config{
		VersionDirectory: versionDir,
		ProjectDirectory: projectDir,
	}

	fmt.Println(`To use Mendix Platform features (mx sync), you need a Personal Access Token (PAT).
The PAT must include the following permissions:
  - mx:mxid3:user-identifiers:uuid:read
  - mx:app:metadata:read
  - mx:modelrepository:repo:read
  - mx:modelrepository:repo:write
  - mx:modelrepository:write

The PAT will be stored as the MX_PAT environment variable — it will NOT be written to the config file.
Press Enter to skip this step. These values can be set later by re-running the CLI or editing the config.`)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("User ID: ")
	userID, _ := reader.ReadString('\n')
	cfg.UserID = strings.TrimSpace(userID)

	fmt.Print("Personal Access Token (PAT): ")
	patBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		fmt.Printf("Warning: could not read PAT securely: %v\n", err)
	} else {
		pat := strings.TrimSpace(string(patBytes))
		if pat != "" {
			if err := persistPAT(pat); err != nil {
				fmt.Printf("Warning: could not persist MX_PAT: %v\n", err)
			}
		}
	}

	return cfg, nil
}

func persistPAT(pat string) error {
	switch runtime.GOOS {
	case "windows":
		if err := exec.Command("setx", "MX_PAT", pat).Run(); err != nil {
			return fmt.Errorf("setx failed: %w", err)
		}
		fmt.Println("MX_PAT has been set as a user environment variable.")
		fmt.Println("You must open a new terminal for it to take effect.")
	default:
		fmt.Printf("\nTo persist MX_PAT, add the following to your ~/.bashrc or ~/.zshrc:\n  export MX_PAT=%s\n", pat)
	}
	return nil
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

func (c *Config) SetApps(apps []App) error {
	c.Apps = apps
	return c.save()
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
