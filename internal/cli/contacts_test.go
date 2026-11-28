package cli

import (
	"reflect"
	"strings"
	"testing"
)

func TestContactsListFirstPage(t *testing.T) {
	srv := fakeDigisac(t, fakeToken)

	stdout, stderr, err := run(t, envCredentials(srv), "", "contacts", "list")
	if err != nil {
		t.Fatalf("contacts list: %v", err)
	}
	want := "ID  NAME          NUMBER         SERVICE\n" +
		"c1  Ada Lovelace  5511999990001  s1\n" +
		"c2  Alan Turing   5511999990002  s1\n"
	if stdout != want {
		t.Errorf("stdout =\n%s\nwant\n%s", stdout, want)
	}
	if !strings.Contains(stderr, "2 contacts\n") {
		t.Errorf("stderr = %q, want it to report 2 contacts", stderr)
	}
}

func TestContactsListAll(t *testing.T) {
	srv := fakeDigisac(t, fakeToken)

	stdout, stderr, err := run(t, envCredentials(srv), "", "contacts", "list", "--all")
	if err != nil {
		t.Fatalf("contacts list --all: %v", err)
	}
	want := "ID  NAME          NUMBER         SERVICE\n" +
		"c1  Ada Lovelace  5511999990001  s1\n" +
		"c2  Alan Turing   5511999990002  s1\n" +
		"c3  Grace Hopper  5511999990003  s2\n"
	if stdout != want {
		t.Errorf("stdout =\n%s\nwant\n%s", stdout, want)
	}
	if !strings.Contains(stderr, "3 contacts\n") {
		t.Errorf("stderr = %q, want it to report 3 contacts", stderr)
	}
	if got, want := srv.requestedContactPages(), []string{"1", "2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("requested pages %v, want %v (stop at lastPage)", got, want)
	}
}

func TestContactsListAllStopsOnEmptyPage(t *testing.T) {
	// lastPage claims 5 pages, but data ends on page 1: page 2 comes back
	// empty, so --all must stop there instead of asking for page 3.
	srv := fakeDigisacWithContacts(t, fakeToken, map[string]string{
		"1": `{"data":[` +
			`{"id":"c1","name":"Ada Lovelace","serviceId":"s1","data":{"number":"5511999990001"}}` +
			`],"total":1,"lastPage":5}`,
		"2": `{"data":[],"total":1,"lastPage":5}`,
	})

	stdout, _, err := run(t, envCredentials(srv), "", "contacts", "list", "--all")
	if err != nil {
		t.Fatalf("contacts list --all: %v", err)
	}
	want := "ID  NAME          NUMBER         SERVICE\n" +
		"c1  Ada Lovelace  5511999990001  s1\n"
	if stdout != want {
		t.Errorf("stdout =\n%s\nwant\n%s", stdout, want)
	}
	if got, want := srv.requestedContactPages(), []string{"1", "2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("requested pages %v, want %v (stop on empty page)", got, want)
	}
}

func TestContactsListJSONIsRawAPIObjects(t *testing.T) {
	// Each contact carries fields the CLI does not model, keys out of
	// alphabetical order, non-ASCII text and HTML-sensitive characters. -o
	// json must print them as the API sent them, except that <, > and & are
	// escaped as \u003c, \u003e and \u0026.
	srv := fakeDigisacWithContacts(t, fakeToken, map[string]string{
		"1": `{"data":[` +
			`{"name":"José Ñúñez","id":"c1","serviceId":"s1","tags":["vip","São Paulo"],"note":"Tom & Jerry <tj@example.com>","data":{"number":"5511999990001","extra":true}},` +
			`{"zeta":1,"id":"c2","name":"Zoë 日本","serviceId":"s2","data":{"number":"5511999990002"},"alpha":null}` +
			`],"total":2,"lastPage":1}`,
	})

	stdout, stderr, err := run(t, envCredentials(srv), "", "contacts", "list", "-o", "json")
	if err != nil {
		t.Fatalf("contacts list -o json: %v", err)
	}
	want := `[
  {
    "name": "José Ñúñez",
    "id": "c1",
    "serviceId": "s1",
    "tags": [
      "vip",
      "São Paulo"
    ],
    "note": "Tom \u0026 Jerry \u003ctj@example.com\u003e",
    "data": {
      "number": "5511999990001",
      "extra": true
    }
  },
  {
    "zeta": 1,
    "id": "c2",
    "name": "Zoë 日本",
    "serviceId": "s2",
    "data": {
      "number": "5511999990002"
    },
    "alpha": null
  }
]
`
	if stdout != want {
		t.Errorf("stdout =\n%s\nwant\n%s", stdout, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestContactsListEmptyTable(t *testing.T) {
	srv := fakeDigisacWithContacts(t, fakeToken, map[string]string{
		"1": `{"data":[],"total":0,"lastPage":1}`,
	})

	stdout, stderr, err := run(t, envCredentials(srv), "", "contacts", "list")
	if err != nil {
		t.Fatalf("contacts list: %v", err)
	}
	if want := "ID  NAME  NUMBER  SERVICE\n"; stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
	if want := "0 contacts\n"; stderr != want {
		t.Errorf("stderr = %q, want %q", stderr, want)
	}
}

func TestContactsListEmptyJSON(t *testing.T) {
	srv := fakeDigisacWithContacts(t, fakeToken, map[string]string{
		"1": `{"data":[],"total":0,"lastPage":1}`,
	})

	stdout, stderr, err := run(t, envCredentials(srv), "", "contacts", "list", "-o", "json")
	if err != nil {
		t.Fatalf("contacts list -o json: %v", err)
	}
	if want := "[]\n"; stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestContactsListRejectsUnknownOutputFormat(t *testing.T) {
	srv := fakeDigisac(t, fakeToken)

	stdout, _, err := run(t, envCredentials(srv), "", "contacts", "list", "-o", "yaml")
	if err == nil {
		t.Fatal("contacts list -o yaml succeeded, want an error")
	}
	if want := `unknown output format "yaml" (want table or json)`; !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to contain %q", err, want)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
}
