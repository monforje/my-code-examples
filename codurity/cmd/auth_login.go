package cmd

import (
	"fmt"

	"codurity/internal/auth"
	"codurity/internal/build"

	"github.com/spf13/cobra"
)

var authLoginCmd = &cobra.Command{
	Use:           "login",
	Short:         "Authenticate via device code",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		fmt.Fprintln(out, banner)

		client := auth.NewClient(build.AuthAPIURL)
		_, err := auth.Login(cmd.Context(), client, out)
		return err
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
}
