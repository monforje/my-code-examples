package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	dirName  = "codurity"
	fileName = "auth.json"
	permDir  = 0o700
	permFile = 0o600
)

// Auth — содержимое auth.json.
type Auth struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// configDir возвращает директорию конфигурации Codurity:
//   - Linux/macOS: $XDG_CONFIG_HOME/codurity или ~/.config/codurity
//   - Windows:     %APPDATA%/Codurity
func configDir() (string, error) {
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA is not set")
		}
		return filepath.Join(appData, "Codurity"), nil
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, dirName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	return filepath.Join(home, ".config", dirName), nil
}

func filePath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fileName), nil
}

// Save записывает auth.json с правами 0600.
func (a *Auth) Save() error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, permDir); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal auth: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, fileName), data, permFile); err != nil {
		return fmt.Errorf("write auth: %w", err)
	}
	return nil
}

// Load читает auth.json.
func Load() (*Auth, error) {
	path, err := filePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read auth: %w", err)
	}
	var a Auth
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("unmarshal auth: %w", err)
	}
	return &a, nil
}

// Exists сообщает, существует ли auth.json.
func Exists() (bool, error) {
	path, err := filePath()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Delete удаляет auth.json (no-op если файла нет).
func Delete() error {
	path, err := filePath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete auth: %w", err)
	}
	return nil
}

// IsExpired сообщает, истёк ли access token.
func (a *Auth) IsExpired() bool {
	return time.Now().After(a.ExpiresAt)
}
