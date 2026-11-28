package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/spf13/cobra"

	"github.com/mguilarducci/mf-digisac/internal/digisac"
)

func newMessagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "messages",
		Aliases: []string{"message"},
		Short:   "Manage messages",
		Long:    `Commands for listing Digisac messages.`,
		RunE:    func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	cmd.AddCommand(newMessagesListCmd())
	return cmd
}

func newMessagesListCmd() *cobra.Command {
	var p digisac.ListMessagesParams
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List one contact's messages, newest first",
		Example: `  # Newest page of a contact's messages as a table
  digisac messages list --contact <id>

  # Every page, as JSON
  digisac messages list --contact <id> --all -o json`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(p.ContactID) == "" {
				return errors.New("--contact must not be empty")
			}
			format, err := listFormat(cmd)
			if err != nil {
				return err
			}
			return runMessagesList(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), format, p)
		},
	}
	cmd.Flags().StringVar(&p.ContactID, "contact", "", "ID of the contact whose messages to list (required)")
	bindPageFlags(cmd.Flags(), &p.PageParams)
	if err := cmd.MarkFlagRequired("contact"); err != nil {
		panic(err) // the flag is defined just above
	}
	return cmd
}

// messagesView shows one message per row and its raw API object as JSON.
var messagesView = listView[digisac.Message]{
	noun:    "messages",
	headers: []string{"ID", "TIME", "FROM", "TYPE", "TEXT"},
	row: func(m digisac.Message) []string {
		return []string{m.ID, m.Timestamp, messageSender(m), string(m.Type), messageText(m)}
	},
	raw: func(m digisac.Message) json.RawMessage { return m.Raw },
}

// messageSender names who sent m: "system" for internal entries, else "bot",
// "me" or "contact", in that order of precedence.
func messageSender(m digisac.Message) string {
	switch {
	case m.Type == digisac.MessageTypeTicket || m.Type == digisac.MessageTypeComment:
		return "system"
	case m.IsFromBot:
		return "bot"
	case m.IsFromMe:
		return "me"
	default:
		return "contact"
	}
}

// oneLine returns s with each CRLF and each other control character (line
// breaks, tabs, escapes) replaced by a space, so it fits one table cell: the
// tabwriter splits cells on tabs, and the others would break the row.
func oneLine(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, strings.ReplaceAll(s, "\r\n", "\n"))
}

// messageText is what the TEXT column shows for m: [deleted] for a deleted
// message, [<type>] for one whose text is empty or only whitespace, otherwise
// its full text on one line (see oneLine).
func messageText(m digisac.Message) string {
	if m.DeletedAt != "" {
		return "[deleted]"
	}
	if text := oneLine(m.Text); strings.TrimSpace(text) != "" {
		return text
	}
	return "[" + string(m.Type) + "]"
}

func runMessagesList(ctx context.Context, stdout, stderr io.Writer, format outputFormat, p digisac.ListMessagesParams) error {
	c, err := digisac.New()
	if err != nil {
		return err //nolint:wrapcheck // digisac.New's error is already the user-facing message ("not authenticated: run ...")
	}

	messages, err := c.ListMessages(ctx, p)
	if err != nil {
		return fmt.Errorf("list messages: %w", err)
	}
	return renderList(stdout, stderr, format, messages, messagesView)
}
