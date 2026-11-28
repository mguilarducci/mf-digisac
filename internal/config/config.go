// Package config stores the saved login in
// $XDG_CONFIG_HOME/digisac/config.json: where the file lives, reading it,
// writing it owner-only, and deleting it. Every credential rule (env
// overrides, URL normalization, "not authenticated") lives in the auth
// package, which is the only caller.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

// Path returns the config file location.
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		dir, err = xdg, nil
	}
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "digisac", "config.json"), nil
}

// Read returns the saved config file's contents. A missing file gives an
// empty Config and no error.
func Read() (*Config, error) {
	cfg := &Config{}
	path, err := Path()
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(path)
	switch {
	case err == nil:
		if err := json.Unmarshal(raw, cfg); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
	case !errors.Is(err, os.ErrNotExist):
		return nil, err
	}
	return cfg, nil
}

// Save writes the config file with owner-only permissions.
func Save(cfg *Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o600)
}

// Delete removes the config file. A missing file is not an error.
func Delete() error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
