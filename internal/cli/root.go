package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mguilarducci/mf-digisac/internal/config"
	"github.com/mguilarducci/mf-digisac/internal/digisac"
)

// Version will be linked by an ldflag during build.
var Version = "v0.1.0-dev"

var rootCmd = &cobra.Command{
	Use:           "digisac",
	Short:         "digisac is a CLI for the Digisac API.",
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
	Run:           func(cmd *cobra.Command, args []string) { _ = cmd.Help() },
}

// Execute adds all child commands to the root command and runs it.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func init() {
	// When run as an mf plugin (`mf digisac ...`), show that in help/usage.
	if os.Getenv("MF_PLUGIN_NAME") != "" {
		rootCmd.Annotations = map[string]string{cobra.CommandDisplayNameAnnotation: "mf digisac"}
	}
	rootCmd.PersistentFlags().StringP("output", "o", "table", "Output format: table or json")
}

func isJSONOutput(cmd *cobra.Command) bool {
	out, _ := cmd.Flags().GetString("output")
	return out == "json"
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
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
