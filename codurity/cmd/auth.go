package cmd

import "github.com/spf13/cobra"

var authCmd = &cobra.Command{
	Use:           "auth",
	Short:         "Manage authentication",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.AddCommand(authCmd)
}
