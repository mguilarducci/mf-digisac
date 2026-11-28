package auth

import (
	"os"
	"path/filepath"
	"testing"
)

// isolate points the config at a temp dir and clears the env overrides.
// Tests that call it must not use t.Parallel.
func isolate(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("DIGISAC_URL", "")
	t.Setenv("DIGISAC_TOKEN", "")
	return dir
}

// writeSavedLogin writes raw as the saved config file under dir, the
// XDG_CONFIG_HOME set by isolate.
func writeSavedLogin(t *testing.T, dir, raw string) {
	t.Helper()
	path := filepath.Join(dir, "digisac", "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
}

const notAuthenticated = "not authenticated: run `digisac auth login` or set DIGISAC_URL and DIGISAC_TOKEN"

func TestCurrentNotAuthenticated(t *testing.T) {
	cases := []struct {
		name           string
		envURL, envTok string
	}{
		{"no file, no env", "", ""},
		{"only url", "https://acme.digisac.io", ""},
		{"only token", "", "tok"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			isolate(t)
			t.Setenv("DIGISAC_URL", c.envURL)
			t.Setenv("DIGISAC_TOKEN", c.envTok)

			_, err := Current()
			if err == nil || err.Error() != notAuthenticated {
				t.Errorf("Current() error = %v, want %q", err, notAuthenticated)
			}
		})
	}
}

func TestCurrentReadsSavedLoginUnderXDGConfigHome(t *testing.T) {
	dir := isolate(t)
	writeSavedLogin(t, dir, `{"url":"https://acme.digisac.io","token":"tok"}`)

	got, err := Current()
	if err != nil {
		t.Fatal(err)
	}
	if want := (Credentials{URL: "https://acme.digisac.io", Token: "tok"}); got != want {
		t.Errorf("Current() = %+v, want %+v", got, want)
	}
}
