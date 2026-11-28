package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// run builds a fresh root command, executes it with argv and stdin, and
// returns what it wrote to stdout and stderr plus the returned error. It points
// XDG_CONFIG_HOME at a temp dir and empties DIGISAC_URL, DIGISAC_TOKEN and
// MF_PLUGIN_NAME, then applies env on top. Callers must not use t.Parallel.
func run(t *testing.T, env map[string]string, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("DIGISAC_URL", "")
	t.Setenv("DIGISAC_TOKEN", "")
	t.Setenv("MF_PLUGIN_NAME", "")
	for k, v := range env {
		t.Setenv(k, v)
	}

	var out, errOut bytes.Buffer
	cmd := NewRootCmd()
	cmd.SetArgs(args)
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	err = cmd.Execute()
	return out.String(), errOut.String(), err
}

const fakeToken = "tok_abcdefghijklmnop"

// contactPages are the fake API's /api/v1/contacts responses, keyed by the
// page query parameter. The listing spans two pages.
var contactPages = map[string]string{
	"1": `{"data":[` +
		`{"id":"c1","name":"Ada Lovelace","serviceId":"s1","data":{"number":"5511999990001"}},` +
		`{"id":"c2","name":"Alan Turing","serviceId":"s1","data":{"number":"5511999990002"}}` +
		`],"total":3,"lastPage":2}`,
	"2": `{"data":[` +
		`{"id":"c3","name":"Grace Hopper","serviceId":"s2","data":{"number":"5511999990003"}}` +
		`],"total":3,"lastPage":2}`,
}

// fakeDigisac serves the parts of the Digisac API the CLI uses. Requests
// without "Authorization: Bearer <token>" get 401.
func fakeDigisac(t *testing.T, token string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"u1","name":"Ada Lovelace","email":"ada@example.com"}`))
	})
	mux.HandleFunc("GET /api/v1/contacts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(contactPages[r.URL.Query().Get("page")]))
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			http.Error(w, `{"message":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		mux.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// envCredentials points DIGISAC_URL and DIGISAC_TOKEN at srv.
func envCredentials(srv *httptest.Server) map[string]string {
	return map[string]string{"DIGISAC_URL": srv.URL + "/api/v1", "DIGISAC_TOKEN": fakeToken}
}

func TestAuthStatusTable(t *testing.T) {
	srv := fakeDigisac(t, fakeToken)

	stdout, _, err := run(t, envCredentials(srv), "", "auth", "status")
	if err != nil {
		t.Fatalf("auth status: %v", err)
	}
	want := "URL:   " + srv.URL + "\n" +
		"User:  Ada Lovelace <ada@example.com>\n" +
		"Token: tok_************mnop\n"
	if stdout != want {
		t.Errorf("stdout =\n%s\nwant\n%s", stdout, want)
	}
}

func TestAuthStatusJSON(t *testing.T) {
	srv := fakeDigisac(t, fakeToken)

	stdout, _, err := run(t, envCredentials(srv), "", "auth", "status", "-o", "json")
	if err != nil {
		t.Fatalf("auth status -o json: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("decode stdout %q: %v", stdout, err)
	}
	want := map[string]any{"id": "u1", "name": "Ada Lovelace", "email": "ada@example.com"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestAuthStatusNotAuthenticated(t *testing.T) {
	_, _, err := run(t, nil, "", "auth", "status")
	if err == nil || !strings.HasPrefix(err.Error(), "not authenticated") {
		t.Fatalf("err = %v, want not authenticated error", err)
	}
}

// configPath returns the config file location under the XDG config dir
// configHome. Pass os.Getenv("XDG_CONFIG_HOME") for the dir run set up.
func configPath(configHome string) string {
	return filepath.Join(configHome, "digisac", "config.json")
}

func TestAuthLoginSavesVerifiedToken(t *testing.T) {
	srv := fakeDigisac(t, fakeToken)

	_, stderr, err := run(t, nil, fakeToken+"\n", "auth", "login", "--url", srv.URL+"/api/v1")
	if err != nil {
		t.Fatalf("auth login: %v", err)
	}
	path := configPath(os.Getenv("XDG_CONFIG_HOME"))
	wantLine := "Logged in to " + srv.URL + " (saved to " + path + ")\n"
	if !strings.HasSuffix(stderr, wantLine) {
		t.Errorf("stderr = %q, want it to end with %q", stderr, wantLine)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var got map[string]string
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode config %q: %v", raw, err)
	}
	want := map[string]string{"url": srv.URL, "token": fakeToken}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("config = %v, want %v", got, want)
	}
}

func TestAuthLoginRejectedTokenSavesNothing(t *testing.T) {
	srv := fakeDigisac(t, fakeToken)

	_, _, err := run(t, nil, "tok_wrong\n", "auth", "login", "--url", srv.URL+"/api/v1")
	if err == nil {
		t.Fatal("auth login with a rejected token succeeded, want error")
	}
	if _, statErr := os.Stat(configPath(os.Getenv("XDG_CONFIG_HOME"))); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("config file exists after rejected login (stat err = %v), want none", statErr)
	}
}
