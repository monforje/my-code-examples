// Package config
package config

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/caarlos0/env/v11"
)

type AppEnv string

const (
	AppEnvLocal AppEnv = "local"
	AppEnvProd  AppEnv = "prod"
)

type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

type Config struct {
	Server     ServerConfig
	PG         PGConfig
	NATS       NATSConfig
	SMTP       SMTPConfig
	Redis      RedisConfig
	Features   FeaturesConfig
	Logger     LoggerConfig
	HTTPClient HTTPClientConfig
}

// HTTPClientConfig - конфигурация исходящих HTTP-клиентов.
type HTTPClientConfig struct {
	ReportsClient ReportsClientConfig `envPrefix:"REPORTS_CLIENT_"`
}

// ReportsClientConfig - конфигурация клиента для отправки CI-отчётов в tasks.
type ReportsClientConfig struct {
	Token   string `env:"TOKEN" envDefault:""`
	BaseURL string `env:"BASE_URL,required"`
}

type ServerConfig struct {
	Port string     `env:"SERVER_PORT" envDefault:"8080"`
	CORS CORSConfig `envPrefix:"CORS_"`
}

type CORSConfig struct {
	Origins       []string `env:"ORIGINS" env-separator:"," envDefault:"*"`
	Methods       []string `env:"METHODS" env-separator:"," envDefault:"GET,POST,PUT,PATCH,DELETE,OPTIONS"`
	Headers       []string `env:"HEADERS" env-separator:"," envDefault:"Authorization,Content-Type"`
	ExposeHeaders []string `env:"EXPOSE_HEADERS" env-separator:","`
}

type SMTPConfig struct {
	Host     string `env:"SMTP_HOST,required"`
	Port     int    `env:"SMTP_PORT" envDefault:"465"`
	Username string `env:"SMTP_USERNAME,required"`
	Password string `env:"SMTP_PASSWORD,required"`
	From     string `env:"SMTP_FROM,required"`
	FromName string `env:"SMTP_FROM_NAME" envDefault:"Codurity"`
}

type PGConfig struct {
	Host     string `env:"POSTGRES_HOST,required"`
	Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
	User     string `env:"POSTGRES_USER,required"`
	Password string `env:"POSTGRES_PASSWORD,required"`
	DB       string `env:"POSTGRES_DB,required"`
	SSLMode  string `env:"POSTGRES_SSL_MODE" envDefault:"disable"`
}

func (cfg *PGConfig) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   net.JoinHostPort(cfg.Host, cfg.Port),
		Path:   cfg.DB,
	}
	q := u.Query()
	q.Set("sslmode", cfg.SSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST,required"`
	Port     string `env:"REDIS_PORT" envDefault:"6379"`
	Password string `env:"REDIS_PASSWORD,required"`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
}

func (cfg *RedisConfig) Addr() string {
	return net.JoinHostPort(cfg.Host, cfg.Port)
}

type NATSConfig struct {
	Host     string `env:"NATS_HOST"`
	Port     string `env:"NATS_PORT" envDefault:"4222"`
	User     string `env:"NATS_USER"`
	Password string `env:"NATS_PASSWORD"`
}

func (cfg *NATSConfig) URL() string {
	return (&url.URL{
		Scheme: "nats",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   net.JoinHostPort(cfg.Host, cfg.Port),
	}).String()
}

type FeaturesConfig struct {
	CIReportCacheTTL time.Duration `env:"CI_REPORT_CACHE_TTL" envDefault:"24h"`
}

type LoggerConfig struct {
	Level  slog.Level
	Format Format
	Output io.Writer
}

func ParseAppEnv(value string) (AppEnv, error) {
	if value == "" {
		return "", fmt.Errorf("APP_ENV is not set")
	}

	appEnv := AppEnv(value)
	switch appEnv {
	case AppEnvLocal, AppEnvProd:
		return appEnv, nil
	default:
		return "", fmt.Errorf("unsupported APP_ENV %q", value)
	}
}

func NewLoggerConfig(appEnv AppEnv) *LoggerConfig {
	switch appEnv {
	case AppEnvLocal:
		return &LoggerConfig{
			Level:  slog.LevelInfo,
			Format: FormatText,
			Output: os.Stdout,
		}
	case AppEnvProd:
		return &LoggerConfig{
			Level:  slog.LevelError,
			Format: FormatJSON,
			Output: os.Stdout,
		}
	default:
		panic(fmt.Sprintf("unsupported APP_ENV %q", appEnv))
	}
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}

	return &cfg, nil
}

func (cfg *Config) Validate() error {
	port, err := strconv.Atoi(cfg.Server.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid SERVER_PORT: %q", cfg.Server.Port)
	}

	if cfg.Redis.DB < 0 || cfg.Redis.DB > 15 {
		return fmt.Errorf("invalid REDIS_DB: %d (must be 0-15)", cfg.Redis.DB)
	}

	if cfg.Features.CIReportCacheTTL <= 0 {
		return fmt.Errorf("CI_REPORT_CACHE_TTL must be positive")
	}

	return nil
}
