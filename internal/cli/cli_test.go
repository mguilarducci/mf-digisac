package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
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

// Paths of the fake API's endpoints.
const (
	mePath       = "/api/v1/me"
	contactsPath = "/api/v1/contacts"
	messagesPath = "/api/v1/messages"
)

// contactPages are the fake API's contacts responses, keyed by the page query
// parameter. The listing spans two pages.
var contactPages = map[string]string{
	"1": `{"data":[` +
		`{"id":"c1","name":"Ada Lovelace","serviceId":"s1","data":{"number":"5511999990001"}},` +
		`{"id":"c2","name":"Alan Turing","serviceId":"s1","data":{"number":"5511999990002"}}` +
		`],"total":3,"lastPage":2}`,
	"2": `{"data":[` +
		`{"id":"c3","name":"Grace Hopper","serviceId":"s2","data":{"number":"5511999990003"}}` +
		`],"total":3,"lastPage":2}`,
}

// messagePages are the fake API's messages responses, keyed by the page query
// parameter. The listing spans two pages, newest first.
var messagePages = map[string]string{
	"1": `{"data":[` +
		`{"id":"m2","timestamp":"2026-11-20T10:05:00.000Z","isFromMe":true,"type":"chat","text":"Hi Ada, how can I help?"},` +
		`{"id":"m1","timestamp":"2026-11-20T10:00:00.000Z","isFromMe":false,"type":"chat","text":"Hello"}` +
		`],"total":3,"lastPage":2}`,
	"2": `{"data":[` +
		`{"id":"m0","timestamp":"2026-11-19T18:30:00.000Z","isFromMe":false,"type":"image","text":""}` +
		`],"total":3,"lastPage":2}`,
}

// routes maps each path of the fake API to its responses, keyed by the page
// query parameter; a request without one is answered from the "" key.
type routes map[string]map[string]string

// defaultRoutes are the endpoints the fake API serves unless a test overrides
// them. Adding an endpoint is one entry here.
var defaultRoutes = routes{
	mePath:       {"": `{"id":"u1","name":"Ada Lovelace","email":"ada@example.com"}`},
	contactsPath: contactPages,
	messagesPath: messagePages,
}

// fakeServer is a fake Digisac API serving routes. It records the query of
// every request it answers, per path.
type fakeServer struct {
	*httptest.Server

	routes routes

	mu      sync.Mutex
	queries map[string][]url.Values
}

// requestedQueries returns the queries of the requests to path served so far,
// in order.
func (f *fakeServer) requestedQueries(path string) []url.Values {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]url.Values(nil), f.queries[path]...)
}

// pageNumbers returns the page parameter of each query, in order.
func pageNumbers(queries []url.Values) []string {
	pages := make([]string, len(queries))
	for i, q := range queries {
		pages[i] = q.Get("page")
	}
	return pages
}

// fakeDigisac serves defaultRoutes, with each path in overrides replacing the
// default one. Requests without "Authorization: Bearer <fakeToken>" get 401;
// an unknown path or page gets 404.
func fakeDigisac(t *testing.T, overrides ...routes) *fakeServer {
	t.Helper()
	f := &fakeServer{routes: routes{}, queries: map[string][]url.Values{}}
	for _, r := range append([]routes{defaultRoutes}, overrides...) {
		for path, pages := range r {
			f.routes[path] = pages
		}
	}
	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+fakeToken {
			http.Error(w, `{"message":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		pages, ok := f.routes[r.URL.Path]
		if !ok || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		f.mu.Lock()
		f.queries[r.URL.Path] = append(f.queries[r.URL.Path], q)
		f.mu.Unlock()
		body, ok := pages[q.Get("page")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(body)); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(f.Close)
	return f
}

// envCredentials points DIGISAC_URL and DIGISAC_TOKEN at srv.
func envCredentials(srv *fakeServer) map[string]string {
	return map[string]string{"DIGISAC_URL": srv.URL + "/api/v1", "DIGISAC_TOKEN": fakeToken}
}

func TestAuthStatusTable(t *testing.T) {
	srv := fakeDigisac(t)

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
	srv := fakeDigisac(t)

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

func TestNotAuthenticated(t *testing.T) {
	cases := [][]string{
		{"auth", "status"},
		{"contacts", "list"},
		{"messages", "list", "--contact", "c1"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			_, _, err := run(t, nil, "", args...)
			if err == nil || !strings.HasPrefix(err.Error(), "not authenticated") {
				t.Fatalf("err = %v, want not authenticated error", err)
			}
		})
	}
}

// configPath returns the config file location under the XDG config dir
// configHome. Pass os.Getenv("XDG_CONFIG_HOME") for the dir run set up.
func configPath(configHome string) string {
	return filepath.Join(configHome, "digisac", "config.json")
}

func TestAuthLoginSavesVerifiedToken(t *testing.T) {
	srv := fakeDigisac(t)

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
	srv := fakeDigisac(t)

	_, _, err := run(t, nil, "tok_wrong\n", "auth", "login", "--url", srv.URL+"/api/v1")
	if err == nil {
		t.Fatal("auth login with a rejected token succeeded, want error")
	}
	if _, statErr := os.Stat(configPath(os.Getenv("XDG_CONFIG_HOME"))); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("config file exists after rejected login (stat err = %v), want none", statErr)
	}
}
