package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mx",
	Short: "Mendix Project Manager",
	Long:  `A CLI tool for managing Mendix Studio Pro versions and projects.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
