package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// outputFormat is how a list command renders its items.
type outputFormat int

const (
	formatTable outputFormat = iota
	formatJSON
)

// listFormat parses the -o/--output flag of cmd. It accepts table and json
// and rejects anything else.
func listFormat(cmd *cobra.Command) (outputFormat, error) {
	out, err := cmd.Flags().GetString("output")
	if err != nil {
		return 0, fmt.Errorf("read --output flag: %w", err)
	}
	switch out {
	case "table":
		return formatTable, nil
	case "json":
		return formatJSON, nil
	}
	return 0, fmt.Errorf("unknown output format %q (want table or json)", out)
}

// listView declares how a list command shows its items.
type listView[T any] struct {
	// noun names the items on the count line, e.g. "contacts".
	noun string
	// headers are the table's column headers.
	headers []string
	// row maps one item to its cells, in header order.
	row func(T) []string
	// raw returns one item's raw API object for JSON output.
	raw func(T) json.RawMessage
}

// renderList writes items to stdout as a table (followed by an "N <noun>"
// count line on stderr) or as a JSON array of their raw API objects.
func renderList[T any](stdout, stderr io.Writer, format outputFormat, items []T, view listView[T]) error {
	if format == formatJSON {
		raw := make([]json.RawMessage, len(items))
		for i, it := range items {
			raw[i] = view.raw(it)
		}
		return printJSON(stdout, raw)
	}

	w := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(view.headers, "\t"))
	for _, it := range items {
		fmt.Fprintln(w, strings.Join(view.row(it), "\t"))
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("write table: %w", err)
	}
	fmt.Fprintf(stderr, "%d %s\n", len(items), view.noun)
	return nil
}
