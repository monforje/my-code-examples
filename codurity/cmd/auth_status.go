package cmd

import (
	"fmt"
	"time"

	"codurity/internal/auth"
	"codurity/internal/build"

	"github.com/spf13/cobra"
)

var authStatusCmd = &cobra.Command{
	Use:           "status",
	Short:         "Show authentication status",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		exists, err := auth.Exists()
		if err != nil {
			return fmt.Errorf("auth status: %w", err)
		}
		if !exists {
			fmt.Fprintln(out, "Вы не вошли в систему.")
			return nil
		}

		a, err := auth.Load()
		if err != nil {
			return fmt.Errorf("auth status: %w", err)
		}

		// Если access token истёк — пробуем обновить через refresh token.
		if a.IsExpired() && a.RefreshToken != "" {
			if refreshed, rerr := auth.NewClient(build.AuthAPIURL).Refresh(cmd.Context(), a.RefreshToken); rerr == nil {
				a.AccessToken = refreshed.AccessToken
				a.RefreshToken = refreshed.RefreshToken
				a.ExpiresAt = time.Now().Add(time.Duration(refreshed.ExpiresIn) * time.Second).UTC()
				_ = a.Save()
			}
		}

		fmt.Fprintln(out, "Вы вошли в систему.")
		fmt.Fprintf(out, "Сессия истекает: %s\n", a.ExpiresAt.Format(time.RFC3339))
		if a.IsExpired() {
			fmt.Fprintln(out, "Сессия истекла. Выполните вход заново.")
		}
		return nil
	},
}

func init() {
	authCmd.AddCommand(authStatusCmd)
}
