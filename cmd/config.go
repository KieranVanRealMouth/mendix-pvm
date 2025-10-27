package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"mendix-project-manager/internal/config"
	"mendix-project-manager/internal/models"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Open the configuration file",
	Long:  `Opens the Mendix Project Manager configuration file in the default editor.`,
	Run:   runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting config path: %v\n", err)
		os.Exit(1)
	}

	// Create config file if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg := &models.Config{
			VersionDirectory:   `C:\Program Files\Mendix\`,
			ProjectDirectories: []string{},
		}
		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating config file: %v\n", err)
			os.Exit(1)
		}
	}

	// Open the config file
	var openCmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		openCmd = exec.Command("cmd", "/c", "start", "", configPath)
	case "darwin":
		openCmd = exec.Command("open", configPath)
	default:
		openCmd = exec.Command("xdg-open", configPath)
	}

	if err := openCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error opening config file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Opened config file: %s\n", configPath)
}
