package cmd

import (
	"fmt"
	"strings"

	"codurity/internal/api"
	"codurity/internal/backend"

	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone [owner/repo]",
	Short: "Clone a GitHub repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := ConfigFromCtx(cmd)

		parts := strings.SplitN(args[0], "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid repo format, expected owner/repo")
		}
		owner, name := parts[0], parts[1]

		client := api.NewClient(cfg.APIBaseURL, cfg.GitHubToken, cfg.Timeout)
		scanner := backend.NewScanner(client)

		cloneURL, err := scanner.GetRepoCloneURL(cmd.Context(), owner, name)
		if err != nil {
			return fmt.Errorf("get repo: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Clone URL: %s\n", cloneURL)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}
