package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// savedConfig writes a config file into a fresh XDG config dir and returns the
// env that points run at it, plus the file's path.
func savedConfig(t *testing.T) (env map[string]string, path string) {
	t.Helper()
	dir := t.TempDir()
	path = configPath(dir)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	raw := `{"url":"https://acme.digisac.io","token":"` + fakeToken + `"}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	return map[string]string{"XDG_CONFIG_HOME": dir}, path
}

func TestAuthLogoutRemovesSavedConfig(t *testing.T) {
	env, path := savedConfig(t)

	stdout, stderr, err := run(t, env, "", "auth", "logout")
	if err != nil {
		t.Fatalf("auth logout: %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("config file still exists after logout (stat err = %v)", err)
	}
	if stderr != "Logged out.\n" {
		t.Errorf("stderr = %q, want %q", stderr, "Logged out.\n")
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
}

func TestAuthLogoutNotesTokenStillInEnv(t *testing.T) {
	env, path := savedConfig(t)
	env["DIGISAC_TOKEN"] = fakeToken

	_, stderr, err := run(t, env, "", "auth", "logout")
	if err != nil {
		t.Fatalf("auth logout: %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("config file still exists after logout (stat err = %v)", err)
	}
	want := "Logged out.\nNote: DIGISAC_TOKEN is still set in your environment.\n"
	if stderr != want {
		t.Errorf("stderr = %q, want %q", stderr, want)
	}
}
