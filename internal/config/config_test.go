package config

import (
	"os"
	"runtime"
	"testing"
)

func isolate(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("DIGISAC_URL", "")
	t.Setenv("DIGISAC_TOKEN", "")
	return dir
}

func TestSaveReadRoundTrip(t *testing.T) {
	isolate(t)
	if err := Save(&Config{URL: "https://acme.digisac.io", Token: "tok"}); err != nil {
		t.Fatal(err)
	}

	path, _ := Path()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	// Windows has no Unix permission bits; os.WriteFile's 0o600 maps to 0o666.
	if perm := info.Mode().Perm(); runtime.GOOS != "windows" && perm != 0o600 {
		t.Errorf("config file perm = %o, want 600", perm)
	}

	cfg, err := Read()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.URL != "https://acme.digisac.io" || cfg.Token != "tok" {
		t.Errorf("Read() = %+v", cfg)
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
