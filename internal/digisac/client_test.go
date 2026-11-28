package digisac

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func newTestServer(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return NewWithCredentials(srv.URL, "tok")
}

func TestListContacts(t *testing.T) {
	requests := 0
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/api/v1/contacts" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("page"); got != "2" {
			t.Errorf("page = %q, want 2", got)
		}
		if got := r.URL.Query().Get("perPage"); got != "10" {
			t.Errorf("perPage = %q, want 10", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("Authorization = %q", got)
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"c1","name":"Ana","serviceId":"s1","data":{"number":"5511"},"extra":true}],"total":11,"lastPage":2}`))
	})

	contacts, err := c.ListContacts(context.Background(), ListContactsParams{Page: 2, PerPage: 10})
	if err != nil {
		t.Fatal(err)
	}
	if requests != 1 {
		t.Errorf("requests = %d, want 1", requests)
	}
	if len(contacts) != 1 {
		t.Fatalf("contacts = %+v", contacts)
	}
	ct := contacts[0]
	if ct.ID != "c1" || ct.Name != "Ana" || ct.ServiceID != "s1" || ct.Data.Number != "5511" {
		t.Errorf("contact = %+v", ct)
	}
	if !strings.Contains(string(ct.Raw), `"extra":true`) {
		t.Errorf("Raw should keep unknown fields, got %s", ct.Raw)
	}
}

func TestMe(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/me" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"name":"Ana","email":"ana@x.com"}`))
	})

	me, err := c.Me(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if me["name"] != "Ana" || me["email"] != "ana@x.com" {
		t.Errorf("me = %v", me)
	}
}

func TestErrorStatusIncludesBody(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"invalid token"}`))
	})

	_, err := c.Me(context.Background())
	if err == nil {
		t.Fatal("want error, got nil")
	}
	for _, want := range []string{"401", "invalid token"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should contain %q", err, want)
		}
	}
}

func TestInvalidJSONResponse(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>oops</html>`))
	})

	if _, err := c.ListContacts(context.Background(), ListContactsParams{Page: 1, PerPage: 1}); err == nil {
		t.Error("want error for non-JSON body, got nil")
	}
}

// pagedServer serves contacts from pages (1-based) and records the pages requested.
// lastPage is what every response reports.
func pagedServer(t *testing.T, lastPage int, pages map[int]string) (*Client, *[]string) {
	t.Helper()
	var requested []string
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Query().Get("page")
		requested = append(requested, p)
		n, _ := strconv.Atoi(p)
		body := pages[n]
		fmt.Fprintf(w, `{"data":[%s],"total":0,"lastPage":%d}`, body, lastPage)
	})
	return c, &requested
}

func contactJSON(id string) string { return `{"id":"` + id + `"}` }

func ids(cs []Contact) string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.ID
	}
	return strings.Join(out, ",")
}

func TestListContactsAllStopsAtLastPage(t *testing.T) {
	c, requested := pagedServer(t, 3, map[int]string{
		1: contactJSON("a") + "," + contactJSON("b"),
		2: contactJSON("c"),
		3: contactJSON("d"),
		4: contactJSON("never"),
	})

	contacts, err := c.ListContacts(context.Background(), ListContactsParams{Page: 1, PerPage: 2, All: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := ids(contacts); got != "a,b,c,d" {
		t.Errorf("ids = %q, want a,b,c,d", got)
	}
	if got := strings.Join(*requested, ","); got != "1,2,3" {
		t.Errorf("pages requested = %q, want 1,2,3", got)
	}
}

func TestListContactsAllStopsOnEmptyPageWithoutLastPage(t *testing.T) {
	c, requested := pagedServer(t, 0, map[int]string{
		1: contactJSON("a"),
		2: contactJSON("b"),
		4: contactJSON("never"),
	})

	contacts, err := c.ListContacts(context.Background(), ListContactsParams{Page: 1, PerPage: 1, All: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := ids(contacts); got != "a,b" {
		t.Errorf("ids = %q, want a,b", got)
	}
	if got := strings.Join(*requested, ","); got != "1,2,3" {
		t.Errorf("pages requested = %q, want 1,2,3", got)
	}
}

func TestListContactsAllStartsAtGivenPage(t *testing.T) {
	c, requested := pagedServer(t, 3, map[int]string{
		1: contactJSON("skipped"),
		2: contactJSON("b"),
		3: contactJSON("c"),
	})

	contacts, err := c.ListContacts(context.Background(), ListContactsParams{Page: 2, PerPage: 1, All: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := ids(contacts); got != "b,c" {
		t.Errorf("ids = %q, want b,c", got)
	}
	if got := strings.Join(*requested, ","); got != "2,3" {
		t.Errorf("pages requested = %q, want 2,3", got)
	}
}

func TestListContactsAllErrorOnLaterPage(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "2" {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"a"}],"lastPage":3}`))
	})

	contacts, err := c.ListContacts(context.Background(), ListContactsParams{Page: 1, PerPage: 1, All: true})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.HasPrefix(err.Error(), "page 2: ") {
		t.Errorf("err = %q, want prefix %q", err, "page 2: ")
	}
	if contacts != nil {
		t.Errorf("contacts = %+v, want nil", contacts)
	}
}
