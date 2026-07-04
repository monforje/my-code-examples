package cmd

import (
	"fmt"

	"codurity/internal/auth"
	"codurity/internal/build"

	"github.com/spf13/cobra"
)

var authLogoutCmd = &cobra.Command{
	Use:           "logout",
	Short:         "Log out and remove local credentials",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		// best-effort: инвалидируем сессию на сервере, если есть токен.
		if a, err := auth.Load(); err == nil {
			_ = auth.NewClient(build.AuthAPIURL).Logout(cmd.Context(), a.AccessToken)
		}

		if err := auth.Delete(); err != nil {
			return fmt.Errorf("logout: %w", err)
		}
		fmt.Fprintln(out, "Вы вышли из аккаунта.")
		return nil
	},
}

func init() {
	authCmd.AddCommand(authLogoutCmd)
}
