// Package digisac is a minimal HTTP client for the Digisac REST API.
package digisac

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{baseURL: baseURL, token: token, http: http.DefaultClient}
}

// Page is the paginated envelope returned by list endpoints.
type Page[T any] struct {
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

type ListContactsParams struct {
	Page    int
	PerPage int
}

func (c *Client) ListContacts(ctx context.Context, p ListContactsParams) (*Page[Contact], error) {
	q := url.Values{"page": {fmt.Sprint(p.Page)}, "perPage": {fmt.Sprint(p.PerPage)}}
	var out Page[Contact]
	if err := c.get(ctx, "/api/v1/contacts?"+q.Encode(), &out); err != nil {
		return nil, err
	}
	return &out, nil
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
