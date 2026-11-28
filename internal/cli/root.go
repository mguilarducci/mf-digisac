package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/mguilarducci/mf-digisac/internal/config"
	"github.com/mguilarducci/mf-digisac/internal/digisac"
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
		Run:           func(cmd *cobra.Command, args []string) { _ = cmd.Help() },
	}
	// When run as an mf plugin (`mf digisac ...`), show that in help/usage.
	if os.Getenv("MF_PLUGIN_NAME") != "" {
		cmd.Annotations = map[string]string{cobra.CommandDisplayNameAnnotation: "mf digisac"}
	}
	cmd.PersistentFlags().StringP("output", "o", "table", "Output format: table or json")
	cmd.AddCommand(newVersionCmd(), newAuthCmd(), newContactsCmd())
	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the plugin version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), "mf-digisac "+Version)
		},
	}
}

func isJSONOutput(cmd *cobra.Command) bool {
	out, _ := cmd.Flags().GetString("output")
	return out == "json"
}

func printJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// clientFromConfig builds an API client from the stored config and env vars.
func clientFromConfig() (*digisac.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	if cfg.URL == "" || cfg.Token == "" {
		return nil, fmt.Errorf("not authenticated: run `digisac auth login` or set DIGISAC_URL and DIGISAC_TOKEN")
	}
	return digisac.New(cfg.URL, cfg.Token), nil
}
