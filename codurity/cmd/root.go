package cmd

import (
	"context"
	"fmt"

	"codurity/internal/config"

	"github.com/spf13/cobra"
)

var banner = `
                         ▒███                     ▒███  ▒███              
                         ▒███                     ▒▒▒▒  ▒███               
  ██████    ██████    ███████ ▒███ ▒███  ▒██████  ▒███  ▒█████   ▒███ ▒███
 ███▒▒███  ███▒▒███  ███▒ ▒██ ▒███ ▒███ ▒███▒▒███ ▒███  ▒███▒    ▒███ ▒███ 
▒███ ▒▒▒  ▒███ ▒███ ▒███  ▒██ ▒███ ▒███ ▒███ ▒▒▒  ▒███  ▒███     ▒███ ▒███ 
▒███  ███ ▒███ ▒███ ▒███▒ ▒██ ▒███ ▒███ ▒███      ▒███  ▒███ ███ ▒███ ▒███ 
▒▒██████  ▒▒██████   ▒███████ ▒▒██████▒ ▒███      ▒███  ▒▒█████   ▒███████ 
 ▒▒▒▒▒▒    ▒▒▒▒▒▒     ▒▒▒▒▒▒   ▒▒▒▒▒▒▒  ▒▒▒▒      ▒▒▒▒   ▒▒▒▒▒       ▒▒███ 
                                                                 ███   ▒██ 
                                                                 ▒███▒▒███  
                                                                  ▒▒████▒   
`

type contextKey struct{}

var cfgFile string

var rootCmd = &cobra.Command{
	Use:           "codurity",
	Short:         "Security scanning and repository management tool",
	Long:          "Codurity — CLI tool for security scanning, repository cloning, and API integration.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cmd, cfgFile)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		ctx := context.WithValue(cmd.Context(), contextKey{}, cfg)
		cmd.SetContext(ctx)
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $HOME/.codurity.yaml)")
	rootCmd.PersistentFlags().String("token", "", "GitHub API token")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}

func ConfigFromCtx(cmd *cobra.Command) *config.Config {
	return cmd.Context().Value(contextKey{}).(*config.Config)
}
