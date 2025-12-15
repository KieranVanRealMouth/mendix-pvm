package cmd

import (
	"fmt"
	"mendix-pvm/internal/config"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config [key] [value]",
	Short: "View or edit the config by specifying the key and value.",
	Long:  "View config values by specifying the key, and edit their values by including the value",
	Args:  cobra.MaximumNArgs(2),
	RunE:  runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch len(args) {
	case 0: // show full config
		fmt.Println("Project Directory: " + cfg.ProjectDirectory)
		fmt.Println("Version Directory: " + cfg.VersionDirectory)
		return nil
	case 1: // show value for provided key
		value, err := cfg.Get(args[0])
		if err != nil {
			return err
		}

		fmt.Println(value)
		return nil
	case 2:
		key := args[0]
		value := args[1]

		if err := cfg.Set(key, value); err != nil {
			return err
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("updated %s to %s\n", key, value)
		return nil
	default:
		return fmt.Errorf("an error occurred: too many arguments")
	}
}
