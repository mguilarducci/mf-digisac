package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/mguilarducci/mf-digisac/internal/config"
	"github.com/mguilarducci/mf-digisac/internal/digisac"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage Digisac credentials",
		Long:  `Commands for saving, checking, and removing the Digisac URL and API token.`,
		Run:   func(cmd *cobra.Command, args []string) { _ = cmd.Help() },
	}
	cmd.AddCommand(newAuthLoginCmd(), newAuthStatusCmd(), newAuthLogoutCmd())
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var url string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Save the Digisac URL and API token",
		Long: `Save the Digisac URL and API token to the config file.
The token is always prompted for and read without echo, so it never lands in
shell history. When stdin is not a terminal, the token is read from stdin.`,
		Example: `  # Interactive
  digisac auth login

  # URL as a flag, token prompted
  digisac auth login --url https://acme.digisac.io/api/v1

  # Token from a secret manager, via stdin
  op read op://vault/digisac/token | digisac auth login --url https://acme.digisac.io/api/v1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthLogin(cmd.Context(), url)
		},
	}
	cmd.Flags().StringVar(&url, "url", "", "Digisac URL, with or without /api/v1")
	return cmd
}

func runAuthLogin(ctx context.Context, url string) error {
	in := bufio.NewReader(os.Stdin)

	if url == "" {
		fmt.Fprint(os.Stderr, "Digisac URL (e.g. https://acme.digisac.io/api/v1): ")
		line, _ := in.ReadString('\n')
		url = strings.TrimSpace(line)
	}
	var token string
	if term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprint(os.Stderr, "API token: ")
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return err
		}
		token = string(b)
	} else {
		line, _ := in.ReadString('\n')
		token = line
	}
	token = strings.TrimSpace(token)
	if url == "" || token == "" {
		return fmt.Errorf("url and token are required")
	}

	cfg := &config.Config{URL: config.NormalizeURL(url), Token: token}
	if _, err := digisac.New(cfg.URL, cfg.Token).Me(ctx); err != nil {
		return fmt.Errorf("token check failed: %w", err)
	}
	if err := config.Save(cfg); err != nil {
		return err
	}
	path, _ := config.Path()
	fmt.Fprintf(os.Stderr, "Logged in to %s (saved to %s)\n", cfg.URL, path)
	return nil
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show which account the current credentials belong to",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthStatus(cmd.Context(), cmd.OutOrStdout(), isJSONOutput(cmd))
		},
	}
}

func runAuthStatus(ctx context.Context, out io.Writer, jsonOut bool) error {
	c, err := clientFromConfig()
	if err != nil {
		return err
	}
	me, err := c.Me(ctx)
	if err != nil {
		return err
	}
	if jsonOut {
		return printJSON(out, me)
	}
	cfg, _ := config.Load()
	fmt.Fprintf(out, "URL:   %s\nUser:  %v <%v>\nToken: %s\n", cfg.URL, me["name"], me["email"], mask(cfg.Token))
	return nil
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove the saved credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthLogout()
		},
	}
}

func runAuthLogout() error {
	if err := config.Delete(); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Logged out.")
	if os.Getenv("DIGISAC_TOKEN") != "" {
		fmt.Fprintln(os.Stderr, "Note: DIGISAC_TOKEN is still set in your environment.")
	}
	return nil
}

func mask(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}
