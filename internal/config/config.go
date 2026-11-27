// Package config loads and saves the CLI credentials.
//
// Resolution order: DIGISAC_URL / DIGISAC_TOKEN env vars override the values
// stored by `digisac auth login` in $XDG_CONFIG_HOME/digisac/config.json.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// Load reads the config file (if any) and applies env var overrides.
func Load() (*Config, error) {
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
	if v := os.Getenv("DIGISAC_URL"); v != "" {
		cfg.URL = v
	}
	if v := os.Getenv("DIGISAC_TOKEN"); v != "" {
		cfg.Token = v
	}
	cfg.URL = NormalizeURL(cfg.URL)
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

// NormalizeURL strips trailing slashes and a trailing /api/v1, so both
// https://x.digisac.io and https://x.digisac.io/api/v1 work.
func NormalizeURL(u string) string {
	return strings.TrimSuffix(strings.TrimRight(u, "/"), "/api/v1")
}
