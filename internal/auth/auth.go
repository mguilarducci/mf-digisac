// Package auth owns the credential rules: DIGISAC_URL and DIGISAC_TOKEN
// override the saved login field by field, the URL is normalized, and both
// fields must be set to be authenticated. The config package only stores the
// file.
package auth

import (
	"errors"
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
// returns errors from reading or parsing the saved file as they are.
func Current() (Credentials, error) {
	cfg, err := config.Read()
	if err != nil {
		return Credentials{}, err
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

// normalizeURL strips trailing slashes and a trailing /api/v1, so both
// https://x.digisac.io and https://x.digisac.io/api/v1 work.
func normalizeURL(u string) string {
	return strings.TrimSuffix(strings.TrimRight(u, "/"), "/api/v1")
}
