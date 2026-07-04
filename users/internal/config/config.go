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
	JWT        JWTConfig
	PG         PGConfig
	NATS       NATSConfig
	Storage    StorageConfig
	Logger     LoggerConfig
	Features   FeaturesConfig
	HttpClient HttpClientConfig `envPrefix:"HTTP_CLIENT_"`
}

type HttpClientConfig struct {
	GitAuthClient GitAuthClientConfig `envPrefix:"GIT_AUTH_CLIENT_"`
}

type GitAuthClientConfig struct {
	Token   string `env:"TOKEN,required"`
	BaseURL string `env:"BASE_URL,required"`
}

type StorageConfig struct {
	AvatarDir    string `env:"AVATAR_DIR" envDefault:"./uploads/avatars"`
	AvatarPublic string `env:"AVATAR_PUBLIC" envDefault:"/uploads/avatars"`
}

type ServerConfig struct {
	Port         string     `env:"SERVER_PORT" envDefault:"8080"`
	ServiceToken string     `env:"SERVICE_TOKEN,required"`
	CORS         CORSConfig `envPrefix:"CORS_"`
}

type CORSConfig struct {
	Origins       []string `env:"ORIGINS" env-separator:"," envDefault:"*"`
	Methods       []string `env:"METHODS" env-separator:"," envDefault:"GET,POST,PUT,PATCH,DELETE,OPTIONS"`
	Headers       []string `env:"HEADERS" env-separator:"," envDefault:"Authorization,Content-Type"`
	ExposeHeaders []string `env:"EXPOSE_HEADERS" env-separator:","`
}

type JWTConfig struct {
	Secret string `env:"JWT_SIGNING_KEY,required"`
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

type NATSConfig struct {
	Host     string `env:"NATS_HOST,required"`
	Port     string `env:"NATS_PORT" envDefault:"4222"`
	User     string `env:"NATS_USER,required"`
	Password string `env:"NATS_PASSWORD,required"`
}

func (cfg *NATSConfig) URL() string {
	return (&url.URL{
		Scheme: "nats",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   net.JoinHostPort(cfg.Host, cfg.Port),
	}).String()
}

type FeaturesConfig struct {
	AccessTokenTTL  time.Duration `env:"ACCESS_TOKEN_TTL" envDefault:"15m"`
	RefreshTokenLen int           `env:"REFRESH_TOKEN_LEN" envDefault:"32"`
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

	if len(cfg.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SIGNING_KEY too short: %d chars (min 32)", len(cfg.JWT.Secret))
	}

	if cfg.Features.AccessTokenTTL <= 0 {
		return fmt.Errorf("ACCESS_TOKEN_TTL must be positive")
	}

	return nil
}
