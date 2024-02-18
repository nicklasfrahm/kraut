package main

import (
	"os"

	"github.com/spf13/cobra"
)

var sshCommand = &cobra.Command{
	Use:   "ssh",
	Short: `Perform SSH operations`,
	Long: `This command group allows it to perform operations
via SSH on a server or a group of servers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Help()
		os.Exit(1)
		return nil
	},
}

func init() {
	sshCommand.AddCommand(sshFingerprintCmd)
}
