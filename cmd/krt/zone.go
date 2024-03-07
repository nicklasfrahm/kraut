package main

import (
	"os"

	"github.com/spf13/cobra"
)

var zoneCommand = &cobra.Command{
	Use:   "zone",
	Short: `Manage availability zones`,
	Long: `Bootstrap new availability zones and manage the
lifecycle of existing ones.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Help()
		os.Exit(1)
		return nil
	},
}

func init() {
	zoneCommand.AddCommand(zoneUpCmd)
}