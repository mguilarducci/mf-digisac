package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/mguilarducci/mf-digisac/internal/digisac"
)

var contactsCmd = &cobra.Command{
	Use:     "contacts",
	Aliases: []string{"contact"},
	Short:   "Manage contacts",
	Long:    `Commands for listing Digisac contacts.`,
	Run:     func(cmd *cobra.Command, args []string) { _ = cmd.Help() },
}

var contactsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List contacts",
	Example: `  # First page as a table
  digisac contacts list

  # Every page, as JSON
  digisac contacts list --all -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")
		all, _ := cmd.Flags().GetBool("all")

		c, err := clientFromConfig()
		if err != nil {
			return err
		}

		var contacts []digisac.Contact
		for p := page; ; p++ {
			res, err := c.ListContacts(cmd.Context(), digisac.ListContactsParams{Page: p, PerPage: perPage})
			if err != nil {
				return fmt.Errorf("list contacts: %w", err)
			}
			contacts = append(contacts, res.Data...)
			if !all || len(res.Data) == 0 || (res.LastPage > 0 && p >= res.LastPage) {
				break
			}
		}

		if isJSONOutput(cmd) {
			raw := make([]json.RawMessage, len(contacts))
			for i, ct := range contacts {
				raw[i] = ct.Raw
			}
			return printJSON(raw)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tNUMBER\tSERVICE")
		for _, ct := range contacts {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", ct.ID, ct.Name, ct.Data.Number, ct.ServiceID)
		}
		if err := w.Flush(); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "%d contacts\n", len(contacts))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(contactsCmd)
	contactsCmd.AddCommand(contactsListCmd)

	contactsListCmd.Flags().Int("page", 1, "Page number")
	contactsListCmd.Flags().Int("per-page", 40, "Items per page")
	contactsListCmd.Flags().Bool("all", false, "Fetch every page")
}
