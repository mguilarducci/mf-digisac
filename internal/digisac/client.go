// Package digisac is a minimal HTTP client for the Digisac REST API.
package digisac

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

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
		return nil, err //nolint:wrapcheck // auth.Current's error is already the user-facing message ("not authenticated: run ...")
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
		return fmt.Errorf("decode contact: %w", err)
	}
	c.Raw = append(json.RawMessage(nil), b...)
	return nil
}

// PageParams selects which pages a list call fetches: page Page of PerPage
// items or, with All set, every page from Page through the last page.
type PageParams struct {
	Page    int
	PerPage int
	All     bool
}

// ListContactsParams selects which contacts to list.
type ListContactsParams struct {
	PageParams
}

// ListContacts lists contacts (GET /api/v1/contacts).
func (c *Client) ListContacts(ctx context.Context, p ListContactsParams) ([]Contact, error) {
	return collectPages[Contact](ctx, c, "/api/v1/contacts", url.Values{}, p.PageParams)
}

// MessageType is the kind of a message as the API names it, such as "chat",
// "image", or one of the internal entries below.
type MessageType string

// Internal entries: they record what happened in the conversation rather
// than something the contact or an agent sent.
const (
	MessageTypeTicket  MessageType = "ticket"
	MessageTypeComment MessageType = "comment"
)

// Message keeps the fields the CLI prints; Raw holds the full object.
type Message struct {
	ID        string          `json:"id"`
	Timestamp string          `json:"timestamp"`
	IsFromMe  bool            `json:"isFromMe"`
	Type      MessageType     `json:"type"`
	Text      string          `json:"text"`
	Raw       json.RawMessage `json:"-"`
}

// UnmarshalJSON decodes a message, tolerating a text that is not a string
// (the API owns that shape): such a text is treated as empty.
func (m *Message) UnmarshalJSON(b []byte) error {
	type alias Message
	aux := struct {
		*alias

		Text json.RawMessage `json:"text"`
	}{alias: (*alias)(m)}
	if err := json.Unmarshal(b, &aux); err != nil {
		return fmt.Errorf("decode message: %w", err)
	}
	m.Text = stringOrEmpty(aux.Text)
	m.Raw = append(json.RawMessage(nil), b...)
	return nil
}

// stringOrEmpty returns raw as a string if it is a JSON string, else "".
func stringOrEmpty(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return ""
	}
	return s
}

// ListMessagesParams selects which messages to list: those of ContactID.
type ListMessagesParams struct {
	PageParams

	ContactID string
}

// ListMessages lists one contact's messages, newest first
// (GET /api/v1/messages). The order is always sent explicitly because the
// API's default order is unknown.
func (c *Client) ListMessages(ctx context.Context, p ListMessagesParams) ([]Message, error) {
	q := url.Values{}
	q.Set("where[contactId]", p.ContactID)
	q.Set("order[0][0]", "timestamp")
	q.Set("order[0][1]", "DESC")
	return collectPages[Message](ctx, c, "/api/v1/messages", q, p.PageParams)
}

// collectPages fetches the pages of path that p selects, stopping early on
// an empty page or once the API reports the last page.
func collectPages[T any](ctx context.Context, c *Client, path string, query url.Values, p PageParams) ([]T, error) {
	var items []T
	for n := p.Page; ; n++ {
		q := url.Values{}
		for k, v := range query {
			q[k] = v
		}
		q.Set("page", strconv.Itoa(n))
		q.Set("perPage", strconv.Itoa(p.PerPage))
		var res page[T]
		if err := c.get(ctx, path+"?"+q.Encode(), &res); err != nil {
			return nil, fmt.Errorf("page %d: %w", n, err)
		}
		items = append(items, res.Data...)
		if !p.All || len(res.Data) == 0 || (res.LastPage > 0 && n >= res.LastPage) {
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
		return fmt.Errorf("GET %s: build request: %w", path, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		// The *url.Error from Do already names the method and URL.
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("GET %s: read response: %w", path, err)
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("GET %s: %s: %s", path, resp.Status, raw)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("GET %s: decode response: %w", path, err)
	}
	return nil
}
