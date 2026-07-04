package cmd

import (
	"fmt"

	"codurity/internal/build"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:           "version",
	Short:         "Print the Codurity CLI version",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintf(cmd.OutOrStdout(), "Codurity CLI v%s\n", build.Version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
