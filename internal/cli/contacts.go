package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/mguilarducci/mf-digisac/internal/digisac"
)

func newContactsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "contacts",
		Aliases: []string{"contact"},
		Short:   "Manage contacts",
		Long:    `Commands for listing Digisac contacts.`,
		Run:     func(cmd *cobra.Command, args []string) { _ = cmd.Help() },
	}
	cmd.AddCommand(newContactsListCmd())
	return cmd
}

func newContactsListCmd() *cobra.Command {
	var p digisac.ListContactsParams
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List contacts",
		Example: `  # First page as a table
  digisac contacts list

  # Every page, as JSON
  digisac contacts list --all -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runContactsList(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), listFormat(cmd), p)
		},
	}
	cmd.Flags().IntVar(&p.Page, "page", 1, "Page number")
	cmd.Flags().IntVar(&p.PerPage, "per-page", 40, "Items per page")
	cmd.Flags().BoolVar(&p.All, "all", false, "Fetch every page")
	return cmd
}

// contactsView shows one contact per row and its raw API object as JSON.
var contactsView = listView[digisac.Contact]{
	noun:    "contacts",
	headers: []string{"ID", "NAME", "NUMBER", "SERVICE"},
	row: func(ct digisac.Contact) []string {
		return []string{ct.ID, ct.Name, ct.Data.Number, ct.ServiceID}
	},
	raw: func(ct digisac.Contact) json.RawMessage { return ct.Raw },
}

func runContactsList(ctx context.Context, stdout, stderr io.Writer, format outputFormat, p digisac.ListContactsParams) error {
	c, err := digisac.New()
	if err != nil {
		return err
	}

	contacts, err := c.ListContacts(ctx, p)
	if err != nil {
		return fmt.Errorf("list contacts: %w", err)
	}
	return renderList(stdout, stderr, format, contacts, contactsView)
}
