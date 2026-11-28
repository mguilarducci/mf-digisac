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

	"github.com/mguilarducci/mf-digisac/internal/auth"
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
			return runAuthLogin(cmd.Context(), cmd.InOrStdin(), cmd.ErrOrStderr(), url)
		},
	}
	cmd.Flags().StringVar(&url, "url", "", "Digisac URL, with or without /api/v1")
	return cmd
}

func runAuthLogin(ctx context.Context, stdin io.Reader, stderr io.Writer, url string) error {
	in := bufio.NewReader(stdin)

	if url == "" {
		fmt.Fprint(stderr, "Digisac URL (e.g. https://acme.digisac.io/api/v1): ")
		line, _ := in.ReadString('\n')
		url = strings.TrimSpace(line)
	}
	var token string
	if f, ok := stdin.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		fmt.Fprint(stderr, "API token: ")
		b, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(stderr)
		if err != nil {
			return err
		}
		token = string(b)
	} else {
		line, _ := in.ReadString('\n')
		token = line
	}
	creds, err := auth.ParseLogin(url, strings.TrimSpace(token))
	if err != nil {
		return err
	}
	if _, err := digisac.NewWithCredentials(creds.URL, creds.Token).Me(ctx); err != nil {
		return fmt.Errorf("token check failed: %w", err)
	}
	path, err := auth.RecordLogin(creds)
	if err != nil {
		return err
	}
	fmt.Fprintf(stderr, "Logged in to %s (saved to %s)\n", creds.URL, path)
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

func runAuthStatus(ctx context.Context, stdout io.Writer, jsonOut bool) error {
	creds, err := auth.Current()
	if err != nil {
		return err
	}
	me, err := digisac.NewWithCredentials(creds.URL, creds.Token).Me(ctx)
	if err != nil {
		return err
	}
	if jsonOut {
		return printJSON(stdout, me)
	}
	fmt.Fprintf(stdout, "URL:   %s\nUser:  %v <%v>\nToken: %s\n", creds.URL, me["name"], me["email"], mask(creds.Token))
	return nil
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove the saved credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthLogout(cmd.Context(), cmd.ErrOrStderr())
		},
	}
}

func runAuthLogout(ctx context.Context, stderr io.Writer) error {
	if err := auth.RemoveLogin(); err != nil {
		return err
	}
	fmt.Fprintln(stderr, "Logged out.")
	if os.Getenv("DIGISAC_TOKEN") != "" {
		fmt.Fprintln(stderr, "Note: DIGISAC_TOKEN is still set in your environment.")
	}
	return nil
}

func mask(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}
