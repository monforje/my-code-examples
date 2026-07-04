package token

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	tokenFileName = "token"
	permDir       = 0o700
	permFile      = 0o600
)

func dirPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	return filepath.Join(home, ".codurity"), nil
}

func Save(token string) error {
	dir, err := dirPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, permDir); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	path := filepath.Join(dir, tokenFileName)
	return os.WriteFile(path, []byte(token), permFile)
}

func Load() (string, error) {
	dir, err := dirPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(dir, tokenFileName))
	if err != nil {
		return "", fmt.Errorf("read token: %w", err)
	}
	return string(data), nil
}

func Delete() error {
	dir, err := dirPath()
	if err != nil {
		return err
	}
	return os.Remove(filepath.Join(dir, tokenFileName))
}
