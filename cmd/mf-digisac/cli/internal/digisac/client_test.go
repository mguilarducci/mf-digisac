package digisac

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return New(srv.URL, "tok")
}

func TestListContacts(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
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

	page, err := c.ListContacts(context.Background(), ListContactsParams{Page: 2, PerPage: 10})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 11 || page.LastPage != 2 || len(page.Data) != 1 {
		t.Fatalf("page = %+v", page)
	}
	ct := page.Data[0]
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
