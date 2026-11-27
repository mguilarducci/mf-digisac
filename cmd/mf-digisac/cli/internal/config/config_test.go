package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	cases := []struct{ in, want string }{
		{"https://acme.digisac.io", "https://acme.digisac.io"},
		{"https://acme.digisac.io/", "https://acme.digisac.io"},
		{"https://acme.digisac.io/api/v1", "https://acme.digisac.io"},
		{"https://acme.digisac.io/api/v1/", "https://acme.digisac.io"},
		{"https://acme.digisac.io/api/v1//", "https://acme.digisac.io"},
		{"", ""},
	}
	for _, c := range cases {
		if got := NormalizeURL(c.in); got != c.want {
			t.Errorf("NormalizeURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// isolate points the config at a temp dir and clears the env overrides.
func isolate(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("DIGISAC_URL", "")
	t.Setenv("DIGISAC_TOKEN", "")
	return dir
}

func TestPathUsesXDGConfigHome(t *testing.T) {
	dir := isolate(t)
	got, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "digisac", "config.json"); got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}
}

func TestLoadWithoutFileIsEmpty(t *testing.T) {
	isolate(t)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.URL != "" || cfg.Token != "" {
		t.Errorf("Load() = %+v, want empty", cfg)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	isolate(t)
	if err := Save(&Config{URL: "https://acme.digisac.io", Token: "tok"}); err != nil {
		t.Fatal(err)
	}

	path, _ := Path()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("config file perm = %o, want 600", perm)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.URL != "https://acme.digisac.io" || cfg.Token != "tok" {
		t.Errorf("Load() = %+v", cfg)
	}
}

func TestLoadEnvOverridesFile(t *testing.T) {
	cases := []struct {
		name               string
		envURL, envToken   string
		wantURL, wantToken string
	}{
		{"no env uses file", "", "", "https://file.digisac.io", "file-tok"},
		{"env token only", "", "env-tok", "https://file.digisac.io", "env-tok"},
		{"env url only, normalized", "https://env.digisac.io/api/v1/", "", "https://env.digisac.io", "file-tok"},
		{"both env", "https://env.digisac.io", "env-tok", "https://env.digisac.io", "env-tok"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			isolate(t)
			if err := Save(&Config{URL: "https://file.digisac.io", Token: "file-tok"}); err != nil {
				t.Fatal(err)
			}
			t.Setenv("DIGISAC_URL", c.envURL)
			t.Setenv("DIGISAC_TOKEN", c.envToken)

			cfg, err := Load()
			if err != nil {
				t.Fatal(err)
			}
			if cfg.URL != c.wantURL || cfg.Token != c.wantToken {
				t.Errorf("Load() = {%q %q}, want {%q %q}", cfg.URL, cfg.Token, c.wantURL, c.wantToken)
			}
		})
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	isolate(t)
	path, _ := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(); err == nil {
		t.Error("Load() with invalid JSON: want error, got nil")
	}
}

func TestDelete(t *testing.T) {
	isolate(t)
	if err := Delete(); err != nil {
		t.Errorf("Delete() without file: %v", err)
	}
	if err := Save(&Config{URL: "u", Token: "t"}); err != nil {
		t.Fatal(err)
	}
	if err := Delete(); err != nil {
		t.Fatal(err)
	}
	path, _ := Path()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("config file still exists after Delete(): %v", err)
	}
}
