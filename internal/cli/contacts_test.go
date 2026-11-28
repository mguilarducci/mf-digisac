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
