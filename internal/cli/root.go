package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// Version will be linked by an ldflag during build.
var Version = "dev"

// NewRootCmd builds the full command tree.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "digisac",
		Short:         "digisac is a CLI for the Digisac API.",
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	// When run as an mf plugin (`mf digisac ...`), show that in help/usage.
	if os.Getenv("MF_PLUGIN_NAME") != "" {
		cmd.Annotations = map[string]string{cobra.CommandDisplayNameAnnotation: "mf digisac"}
	}
	cmd.PersistentFlags().StringP("output", "o", "table", "Output format: table or json")
	cmd.AddCommand(newVersionCmd(), newAuthCmd(), newContactsCmd(), newMessagesCmd())
	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the plugin version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintln(cmd.OutOrStdout(), "mf-digisac "+Version)
		},
	}
}

func isJSONOutput(cmd *cobra.Command) bool {
	out, err := cmd.Flags().GetString("output")
	return err == nil && out == "json"
}

func printJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}
	return nil
}
