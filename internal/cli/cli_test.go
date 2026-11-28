package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

// fakeDigisac serves the parts of the Digisac API the CLI uses. Requests
// without "Authorization: Bearer <token>" get 401.
func fakeDigisac(t *testing.T, token string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"u1","name":"Ada Lovelace","email":"ada@example.com"}`))
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
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %v, want %v", k, got[k], v)
		}
	}
}

func TestAuthStatusNotAuthenticated(t *testing.T) {
	_, _, err := run(t, nil, "", "auth", "status")
	if err == nil || !strings.HasPrefix(err.Error(), "not authenticated") {
		t.Fatalf("err = %v, want not authenticated error", err)
	}
}
