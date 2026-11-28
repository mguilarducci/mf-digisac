// Package digisac is a minimal HTTP client for the Digisac REST API.
package digisac

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/mguilarducci/mf-digisac/internal/auth"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// New returns a client for the current credentials (see auth.Current), or
// the error auth.Current returns, such as "not authenticated".
func New() (*Client, error) {
	creds, err := auth.Current()
	if err != nil {
		return nil, err
	}
	return NewWithCredentials(creds.URL, creds.Token), nil
}

// NewWithCredentials returns a client for an explicit URL and token, such as
// credentials being verified before they are saved.
func NewWithCredentials(baseURL, token string) *Client {
	return &Client{baseURL: baseURL, token: token, http: http.DefaultClient}
}

// page is the paginated envelope returned by list endpoints.
type page[T any] struct {
	Data     []T `json:"data"`
	Total    int `json:"total"`
	LastPage int `json:"lastPage"`
}

// Contact keeps the fields the CLI prints; Raw holds the full object.
type Contact struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ServiceID string `json:"serviceId"`
	Data      struct {
		Number string `json:"number"`
	} `json:"data"`
	Raw json.RawMessage `json:"-"`
}

func (c *Contact) UnmarshalJSON(b []byte) error {
	type alias Contact
	if err := json.Unmarshal(b, (*alias)(c)); err != nil {
		return err
	}
	c.Raw = append(json.RawMessage(nil), b...)
	return nil
}

// ListContactsParams selects which contacts to list. With All set, listing
// starts at Page and continues through the last page.
type ListContactsParams struct {
	Page    int
	PerPage int
	All     bool
}

// ListContacts lists contacts (GET /api/v1/contacts).
func (c *Client) ListContacts(ctx context.Context, p ListContactsParams) ([]Contact, error) {
	return collectPages[Contact](ctx, c, "/api/v1/contacts", url.Values{}, p.Page, p.PerPage, p.All)
}

// collectPages fetches page start of path and, when all is set, every page
// after it until the API reports the last page.
func collectPages[T any](ctx context.Context, c *Client, path string, query url.Values, start, perPage int, all bool) ([]T, error) {
	var items []T
	for n := start; ; n++ {
		q := url.Values{}
		for k, v := range query {
			q[k] = v
		}
		q.Set("page", fmt.Sprint(n))
		q.Set("perPage", fmt.Sprint(perPage))
		var res page[T]
		if err := c.get(ctx, path+"?"+q.Encode(), &res); err != nil {
			return nil, fmt.Errorf("page %d: %w", n, err)
		}
		items = append(items, res.Data...)
		if !all || len(res.Data) == 0 || (res.LastPage > 0 && n >= res.LastPage) {
			return items, nil
		}
	}
}

// Me returns the authenticated user (GET /api/v1/me).
func (c *Client) Me(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.get(ctx, "/api/v1/me", &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("GET %s: %s: %s", path, resp.Status, raw)
	}
	return json.Unmarshal(raw, out)
}
