// Package auth owns the credential rules: DIGISAC_URL and DIGISAC_TOKEN
// override the saved login field by field, the URL is normalized, both
// fields must be set to be authenticated, and what a login needs. It is the
// only writer of the saved login; the config package only stores the file.
package auth

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mguilarducci/mf-digisac/internal/config"
)

// Credentials are a Digisac URL (without /api/v1) and an API token.
type Credentials struct {
	URL   string
	Token string
}

var errNotAuthenticated = errors.New("not authenticated: run `digisac auth login` or set DIGISAC_URL and DIGISAC_TOKEN")

// Current returns the credentials in effect: the saved login with any
// non-empty DIGISAC_URL or DIGISAC_TOKEN applied on top, URL normalized.
// It fails with "not authenticated" when either field ends up empty, and
// wraps errors from reading or parsing the saved file.
func Current() (Credentials, error) {
	cfg, err := config.Read()
	if err != nil {
		return Credentials{}, fmt.Errorf("load saved login: %w", err)
	}
	c := Credentials{URL: cfg.URL, Token: cfg.Token}
	if v := os.Getenv("DIGISAC_URL"); v != "" {
		c.URL = v
	}
	if v := os.Getenv("DIGISAC_TOKEN"); v != "" {
		c.Token = v
	}
	c.URL = normalizeURL(c.URL)
	if c.URL == "" || c.Token == "" {
		return Credentials{}, errNotAuthenticated
	}
	return c, nil
}

// ParseLogin turns the raw URL and token given to auth login into
// credentials: both are required, and the URL is normalized. Trimming
// whitespace is left to the caller.
func ParseLogin(url, token string) (Credentials, error) {
	if url == "" || token == "" {
		return Credentials{}, errors.New("url and token are required")
	}
	return Credentials{URL: normalizeURL(url), Token: token}, nil
}

// RecordLogin saves credentials the caller has already verified and returns
// the path they were saved to.
func RecordLogin(c Credentials) (string, error) {
	if err := config.Save(&config.Config{URL: c.URL, Token: c.Token}); err != nil {
		return "", fmt.Errorf("save login: %w", err)
	}
	path, err := config.Path()
	if err != nil {
		return "", fmt.Errorf("locate saved login: %w", err)
	}
	return path, nil
}

// RemoveLogin deletes the saved login. A missing file is not an error.
func RemoveLogin() error {
	if err := config.Delete(); err != nil {
		return fmt.Errorf("remove saved login: %w", err)
	}
	return nil
}

// normalizeURL strips trailing slashes and a trailing /api/v1, so both
// https://x.digisac.io and https://x.digisac.io/api/v1 work.
func normalizeURL(u string) string {
	return strings.TrimSuffix(strings.TrimRight(u, "/"), "/api/v1")
}
