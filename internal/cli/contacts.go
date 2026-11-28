package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

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
			return runContactsList(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), isJSONOutput(cmd), p)
		},
	}
	cmd.Flags().IntVar(&p.Page, "page", 1, "Page number")
	cmd.Flags().IntVar(&p.PerPage, "per-page", 40, "Items per page")
	cmd.Flags().BoolVar(&p.All, "all", false, "Fetch every page")
	return cmd
}

func runContactsList(ctx context.Context, stdout, stderr io.Writer, jsonOut bool, p digisac.ListContactsParams) error {
	c, err := clientFromConfig()
	if err != nil {
		return err
	}

	contacts, err := c.ListContacts(ctx, p)
	if err != nil {
		return fmt.Errorf("list contacts: %w", err)
	}

	if jsonOut {
		raw := make([]json.RawMessage, len(contacts))
		for i, ct := range contacts {
			raw[i] = ct.Raw
		}
		return printJSON(stdout, raw)
	}

	w := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tNUMBER\tSERVICE")
	for _, ct := range contacts {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", ct.ID, ct.Name, ct.Data.Number, ct.ServiceID)
	}
	if err := w.Flush(); err != nil {
		return err
	}
	fmt.Fprintf(stderr, "%d contacts\n", len(contacts))
	return nil
}
