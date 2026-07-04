package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	GitHubToken string        `mapstructure:"github_token"`
	APIBaseURL  string        `mapstructure:"api_base_url"`
	Timeout     time.Duration `mapstructure:"timeout"`
	OutputDir   string        `mapstructure:"output_dir"`
	Verbose     bool          `mapstructure:"verbose"`
}

func Load(cmd *cobra.Command, cfgFile string) (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("api_base_url", "https://api.github.com")
	v.SetDefault("timeout", "30s")
	v.SetDefault("output_dir", ".")
	v.SetDefault("verbose", false)

	// Config file
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName(".codurity")
		v.SetConfigType("yaml")
		v.AddConfigPath("$HOME")
		v.AddConfigPath(".")
	}

	// Environment variables: CODURITY_GITHUB_TOKEN, CODURITY_API_BASE_URL...
	v.SetEnvPrefix("CODURITY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	// Flags override everything
	if err := v.BindPFlags(cmd.Flags()); err != nil {
		return nil, fmt.Errorf("bind flags: %w", err)
	}
	if err := v.BindPFlags(cmd.PersistentFlags()); err != nil {
		return nil, fmt.Errorf("bind persistent flags: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}
