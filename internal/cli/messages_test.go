package cli

import (
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
