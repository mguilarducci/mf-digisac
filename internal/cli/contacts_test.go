package cli

import (
	"strings"
	"testing"
)

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
