package cli

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestMessagesListFirstPage(t *testing.T) {
	srv := fakeDigisac(t)

	stdout, stderr, err := run(t, envCredentials(srv), "", "messages", "list", "--contact", "c1")
	if err != nil {
		t.Fatalf("messages list: %v", err)
	}
	want := "ID  TIME                      FROM     TYPE  TEXT\n" +
		"m2  2026-11-20T10:05:00.000Z  me       chat  Hi Ada, how can I help?\n" +
		"m1  2026-11-20T10:00:00.000Z  contact  chat  Hello\n"
	if stdout != want {
		t.Errorf("stdout =\n%s\nwant\n%s", stdout, want)
	}
	if !strings.Contains(stderr, "2 messages\n") {
		t.Errorf("stderr = %q, want it to report 2 messages", stderr)
	}
}

func TestMessagesListRequestQuery(t *testing.T) {
	srv := fakeDigisac(t)

	if _, _, err := run(t, envCredentials(srv), "", "messages", "list", "--contact", "c1", "--page", "2", "--per-page", "10"); err != nil {
		t.Fatalf("messages list: %v", err)
	}
	want := []url.Values{{
		"where[contactId]": {"c1"},
		"order[0][0]":      {"timestamp"},
		"order[0][1]":      {"DESC"},
		"page":             {"2"},
		"perPage":          {"10"},
	}}
	if got := srv.requestedMessageQueries(); !reflect.DeepEqual(got, want) {
		t.Errorf("requested queries %v, want %v", got, want)
	}
}

func TestMessagesListRequiresContact(t *testing.T) {
	srv := fakeDigisac(t)

	stdout, _, err := run(t, envCredentials(srv), "", "messages", "list")
	if err == nil {
		t.Fatal("messages list without --contact succeeded, want an error")
	}
	if !strings.Contains(err.Error(), `"contact"`) {
		t.Errorf("error = %q, want it to name the contact flag", err)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if got := srv.requestedMessageQueries(); len(got) != 0 {
		t.Errorf("requested queries %v, want none", got)
	}
}
