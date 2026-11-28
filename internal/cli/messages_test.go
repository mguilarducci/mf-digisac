package cli

import (
	"bytes"
	"encoding/json"
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
	if got := srv.requestedQueries(messagesPath); !reflect.DeepEqual(got, want) {
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
	if got := srv.requestedQueries(messagesPath); len(got) != 0 {
		t.Errorf("requested queries %v, want none", got)
	}
}

func TestMessagesListAll(t *testing.T) {
	srv := fakeDigisac(t)

	stdout, stderr, err := run(t, envCredentials(srv), "", "messages", "list", "--contact", "c1", "--all")
	if err != nil {
		t.Fatalf("messages list --all: %v", err)
	}
	want := "ID  TIME                      FROM     TYPE   TEXT\n" +
		"m2  2026-11-20T10:05:00.000Z  me       chat   Hi Ada, how can I help?\n" +
		"m1  2026-11-20T10:00:00.000Z  contact  chat   Hello\n" +
		"m0  2026-11-19T18:30:00.000Z  contact  image  [image]\n"
	if stdout != want {
		t.Errorf("stdout =\n%s\nwant\n%s", stdout, want)
	}
	if !strings.Contains(stderr, "3 messages\n") {
		t.Errorf("stderr = %q, want it to report 3 messages", stderr)
	}
	if got, want := pageNumbers(srv.requestedQueries(messagesPath)), []string{"1", "2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("requested pages %v, want %v (stop at lastPage)", got, want)
	}
}

func TestMessagesListAllStopsOnEmptyPage(t *testing.T) {
	// lastPage claims 5 pages, but data ends on page 1: page 2 comes back
	// empty, so --all must stop there instead of asking for page 3.
	srv := fakeDigisac(t, routes{messagesPath: {
		"1": `{"data":[` +
			`{"id":"m1","timestamp":"2026-11-20T10:00:00.000Z","isFromMe":false,"type":"chat","text":"Hello"}` +
			`],"total":1,"lastPage":5}`,
		"2": `{"data":[],"total":1,"lastPage":5}`,
	}})

	stdout, _, err := run(t, envCredentials(srv), "", "messages", "list", "--contact", "c1", "--all")
	if err != nil {
		t.Fatalf("messages list --all: %v", err)
	}
	want := "ID  TIME                      FROM     TYPE  TEXT\n" +
		"m1  2026-11-20T10:00:00.000Z  contact  chat  Hello\n"
	if stdout != want {
		t.Errorf("stdout =\n%s\nwant\n%s", stdout, want)
	}
	if got, want := pageNumbers(srv.requestedQueries(messagesPath)), []string{"1", "2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("requested pages %v, want %v (stop on empty page)", got, want)
	}
}

func TestMessagesListJSONIsRawAPIObjects(t *testing.T) {
	// Each message carries fields the CLI does not model and keys out of
	// alphabetical order. -o json must print them as the API sent them.
	srv := fakeDigisac(t, routes{messagesPath: {
		"1": `{"data":[` +
			`{"text":"Olá, tudo bem?","id":"m1","isFromMe":false,"type":"chat","timestamp":"2026-11-20T10:00:00.000Z","contactId":"c1","data":{"ack":3}},` +
			`{"zeta":1,"id":"m2","isFromMe":true,"type":"image","text":null,"timestamp":"2026-11-20T10:05:00.000Z"}` +
			`],"total":2,"lastPage":1}`,
	}})

	stdout, stderr, err := run(t, envCredentials(srv), "", "messages", "list", "--contact", "c1", "-o", "json")
	if err != nil {
		t.Fatalf("messages list -o json: %v", err)
	}
	want := `[
  {
    "text": "Olá, tudo bem?",
    "id": "m1",
    "isFromMe": false,
    "type": "chat",
    "timestamp": "2026-11-20T10:00:00.000Z",
    "contactId": "c1",
    "data": {
      "ack": 3
    }
  },
  {
    "zeta": 1,
    "id": "m2",
    "isFromMe": true,
    "type": "image",
    "text": null,
    "timestamp": "2026-11-20T10:05:00.000Z"
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

func TestMessagesListEmptyJSON(t *testing.T) {
	// An unknown contact, or one without messages, is just an empty listing.
	srv := fakeDigisac(t, routes{messagesPath: {
		"1": `{"data":[],"total":0,"lastPage":1}`,
	}})

	stdout, stderr, err := run(t, envCredentials(srv), "", "messages", "list", "--contact", "unknown", "-o", "json")
	if err != nil {
		t.Fatalf("messages list -o json: %v", err)
	}
	if want := "[]\n"; stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestMessagesListRejectsUnknownOutputFormat(t *testing.T) {
	srv := fakeDigisac(t)

	stdout, _, err := run(t, envCredentials(srv), "", "messages", "list", "--contact", "c1", "-o", "yaml")
	if err == nil {
		t.Fatal("messages list -o yaml succeeded, want an error")
	}
	if want := `unknown output format "yaml" (want table or json)`; !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to contain %q", err, want)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
}

func TestMessagesListRejectsBlankContact(t *testing.T) {
	for _, contact := range []string{"", "   "} {
		srv := fakeDigisac(t)

		stdout, _, err := run(t, envCredentials(srv), "", "messages", "list", "--contact", contact)
		if err == nil {
			t.Fatalf("messages list --contact %q succeeded, want an error", contact)
		}
		if want := "--contact must not be empty"; !strings.Contains(err.Error(), want) {
			t.Errorf("--contact %q: error = %q, want it to contain %q", contact, err, want)
		}
		if stdout != "" {
			t.Errorf("--contact %q: stdout = %q, want empty", contact, stdout)
		}
		if got := srv.requestedQueries(messagesPath); len(got) != 0 {
			t.Errorf("--contact %q: requested queries %v, want none", contact, got)
		}
	}
}

func TestMessagesListDefaultPaging(t *testing.T) {
	srv := fakeDigisac(t)

	if _, _, err := run(t, envCredentials(srv), "", "messages", "list", "--contact", "c1"); err != nil {
		t.Fatalf("messages list: %v", err)
	}
	queries := srv.requestedQueries(messagesPath)
	if len(queries) != 1 {
		t.Fatalf("requested %d queries, want 1", len(queries))
	}
	if got := queries[0].Get("page"); got != "1" {
		t.Errorf("page = %q, want %q", got, "1")
	}
	if got := queries[0].Get("perPage"); got != "40" {
		t.Errorf("perPage = %q, want %q", got, "40")
	}
}

func TestMessagesListNonStringTextCountsAsEmpty(t *testing.T) {
	// The API owns the shape of text; a number, object or bool there must
	// not break the listing. It is shown like an empty text.
	srv := fakeDigisac(t, routes{messagesPath: {
		"1": `{"data":[` +
			`{"id":"m1","timestamp":"2026-11-20T10:00:00.000Z","isFromMe":false,"type":"chat","text":42},` +
			`{"id":"m2","timestamp":"2026-11-20T10:01:00.000Z","isFromMe":false,"type":"chat","text":{"body":"hi"}},` +
			`{"id":"m3","timestamp":"2026-11-20T10:02:00.000Z","isFromMe":false,"type":"chat","text":true}` +
			`],"total":3,"lastPage":1}`,
	}})

	stdout, _, err := run(t, envCredentials(srv), "", "messages", "list", "--contact", "c1")
	if err != nil {
		t.Fatalf("messages list: %v", err)
	}
	want := "ID  TIME                      FROM     TYPE  TEXT\n" +
		"m1  2026-11-20T10:00:00.000Z  contact  chat  [chat]\n" +
		"m2  2026-11-20T10:01:00.000Z  contact  chat  [chat]\n" +
		"m3  2026-11-20T10:02:00.000Z  contact  chat  [chat]\n"
	if stdout != want {
		t.Errorf("stdout =\n%q\nwant\n%q", stdout, want)
	}
}

// everyKindOfMessage is one page holding each kind of message the table must
// keep readable.
const everyKindOfMessage = `{"data":[` +
	`{"id":"m1","timestamp":"2026-11-20T10:00:00.000Z","isFromMe":false,"isFromBot":false,"type":"chat","text":"Hello"},` +
	`{"id":"m2","timestamp":"2026-11-20T10:01:00.000Z","isFromMe":true,"isFromBot":false,"type":"chat","text":"Hi Ada"},` +
	`{"id":"m3","timestamp":"2026-11-20T10:02:00.000Z","isFromMe":true,"isFromBot":true,"type":"chat","text":"Pick an option"},` +
	`{"id":"m4","timestamp":"2026-11-20T10:03:00.000Z","isFromMe":true,"isFromBot":false,"type":"ticket","text":"Ticket opened"},` +
	`{"id":"m5","timestamp":"2026-11-20T10:04:00.000Z","isFromMe":true,"isFromBot":true,"type":"comment","text":"Call back later"},` +
	`{"id":"m6","timestamp":"2026-11-20T10:05:00.000Z","isFromMe":false,"isFromBot":false,"type":"image","text":null},` +
	`{"id":"m7","timestamp":"2026-11-20T10:06:00.000Z","isFromMe":false,"isFromBot":false,"type":"chat","text":"Oops","deletedAt":"2026-11-20T10:07:00.000Z"},` +
	`{"id":"m8","timestamp":"2026-11-20T10:08:00.000Z","isFromMe":false,"isFromBot":false,"type":"chat","text":"First line\nsecond line\r\nthird line, which is long enough that any truncation of the text column would cut it off"}` +
	`],"total":8,"lastPage":1}`

func TestMessagesListEveryKindOfMessage(t *testing.T) {
	srv := fakeDigisacWithMessages(t, map[string]string{"1": everyKindOfMessage})

	stdout, stderr, err := run(t, envCredentials(srv), "", "messages", "list", "--contact", "c1")
	if err != nil {
		t.Fatalf("messages list: %v", err)
	}
	want := "ID  TIME                      FROM     TYPE     TEXT\n" +
		"m1  2026-11-20T10:00:00.000Z  contact  chat     Hello\n" +
		"m2  2026-11-20T10:01:00.000Z  me       chat     Hi Ada\n" +
		"m3  2026-11-20T10:02:00.000Z  bot      chat     Pick an option\n" +
		"m4  2026-11-20T10:03:00.000Z  system   ticket   Ticket opened\n" +
		"m5  2026-11-20T10:04:00.000Z  system   comment  Call back later\n" +
		"m6  2026-11-20T10:05:00.000Z  contact  image    [image]\n" +
		"m7  2026-11-20T10:06:00.000Z  contact  chat     [deleted]\n" +
		"m8  2026-11-20T10:08:00.000Z  contact  chat     First line second line third line, which is long enough that any truncation of the text column would cut it off\n"
	if stdout != want {
		t.Errorf("stdout =\n%s\nwant\n%s", stdout, want)
	}
	if !strings.Contains(stderr, "8 messages\n") {
		t.Errorf("stderr = %q, want it to report 8 messages", stderr)
	}
}

func TestMessagesListEveryKindOfMessageJSONIsRaw(t *testing.T) {
	srv := fakeDigisacWithMessages(t, map[string]string{"1": everyKindOfMessage})

	stdout, _, err := run(t, envCredentials(srv), "", "messages", "list", "--contact", "c1", "-o", "json")
	if err != nil {
		t.Fatalf("messages list -o json: %v", err)
	}
	// The table rewrites FROM and TEXT; JSON must be the API's objects
	// byte for byte, only indented.
	var page struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal([]byte(everyKindOfMessage), &page); err != nil {
		t.Fatal(err)
	}
	var want bytes.Buffer
	if err := json.Indent(&want, page.Data, "", "  "); err != nil {
		t.Fatal(err)
	}
	want.WriteByte('\n')
	if stdout != want.String() {
		t.Errorf("stdout =\n%s\nwant\n%s", stdout, want.String())
	}
}

func TestMessagesListFlattensControlCharacters(t *testing.T) {
	// A tab would split the cell into a new tabwriter column; a bare CR or
	// other control character would garble the terminal. Each becomes a space.
	srv := fakeDigisacWithMessages(t, map[string]string{
		"1": `{"data":[` +
			`{"id":"m1","timestamp":"2026-11-20T10:00:00.000Z","isFromMe":false,"type":"chat","text":"a\tb\rc\u0000d\u001be\u007ff\u0085g"}` +
			`],"total":1,"lastPage":1}`,
	})

	stdout, _, err := run(t, envCredentials(srv), "", "messages", "list", "--contact", "c1")
	if err != nil {
		t.Fatalf("messages list: %v", err)
	}
	want := "ID  TIME                      FROM     TYPE  TEXT\n" +
		"m1  2026-11-20T10:00:00.000Z  contact  chat  a b c d e f g\n"
	if stdout != want {
		t.Errorf("stdout =\n%q\nwant\n%q", stdout, want)
	}
}

func TestMessagesListBlankTextShowsType(t *testing.T) {
	srv := fakeDigisacWithMessages(t, map[string]string{
		"1": `{"data":[` +
			`{"id":"m1","timestamp":"2026-11-20T10:00:00.000Z","isFromMe":false,"type":"chat","text":" \n\t\r\n "}` +
			`],"total":1,"lastPage":1}`,
	})

	stdout, _, err := run(t, envCredentials(srv), "", "messages", "list", "--contact", "c1")
	if err != nil {
		t.Fatalf("messages list: %v", err)
	}
	want := "ID  TIME                      FROM     TYPE  TEXT\n" +
		"m1  2026-11-20T10:00:00.000Z  contact  chat  [chat]\n"
	if stdout != want {
		t.Errorf("stdout =\n%q\nwant\n%q", stdout, want)
	}
}
