package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manu/bb/internal/client"
	"github.com/manu/bb/internal/config"
	"github.com/manu/bb/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func NewAuthCmd(flags *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
	}
	cmd.AddCommand(newAuthLoginCmd(flags))
	cmd.AddCommand(newAuthLogoutCmd(flags))
	cmd.AddCommand(newAuthStatusCmd(flags))
	return cmd
}

func newAuthLoginCmd(flags *GlobalFlags) *cobra.Command {
	var url, token string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Bitbucket Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := flags.Profile
			scanner := bufio.NewScanner(os.Stdin)

			if url == "" {
				fmt.Fprintf(os.Stderr, "Bitbucket Server URL: ")
				if scanner.Scan() {
					url = strings.TrimSpace(scanner.Text())
				}
			}
			if url == "" {
				return fmt.Errorf("URL is required")
			}

			if token == "" {
				fmt.Fprintf(os.Stderr, "Personal access token: ")
				tokenBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					if scanner.Scan() {
						token = strings.TrimSpace(scanner.Text())
					}
				} else {
					token = string(tokenBytes)
					fmt.Fprintln(os.Stderr)
				}
			}
			if token == "" {
				return fmt.Errorf("token is required")
			}

			c := client.New(url, token)
			ctx := context.Background()
			resp, err := c.FindUser(ctx, "", 0, 1)
			if err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}
			var displayName, name string
			if len(resp.Values) > 0 {
				displayName = resp.Values[0].DisplayName
				name = resp.Values[0].Name
			}

			configDir := config.ConfigDir()
			if err := config.SaveProfile(configDir, profile, url, "", ""); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			credsPath := filepath.Join(configDir, "credentials.yaml")
			if err := config.SaveCredentials(credsPath, profile, token); err != nil {
				return fmt.Errorf("failed to save credentials: %w", err)
			}

			fmt.Fprintf(os.Stderr, "✓ Authenticated as %s (%s)\n", displayName, name)
			return nil
		},
	}
	cmd.Flags().StringVar(&url, "url", "", "Bitbucket Server URL")
	cmd.Flags().StringVar(&token, "token", "", "personal access token")
	return cmd
}

func newAuthLogoutCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir := config.ConfigDir()
			credsPath := filepath.Join(configDir, "credentials.yaml")
			if err := config.RemoveCredentials(credsPath, flags.Profile); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Logged out from profile '%s'\n", flags.Profile)
			return nil
		},
	}
}

func newAuthStatusCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			ctx := context.Background()
			resp, err := c.FindUser(ctx, "", 0, 1)
			if err != nil {
				return err
			}
			var userName string
			if len(resp.Values) > 0 {
				userName = resp.Values[0].DisplayName
			}

			type statusInfo struct {
				Profile string `json:"profile"`
				URL     string `json:"url"`
				User    string `json:"user"`
			}
			info := statusInfo{
				Profile: cfg.Profile,
				URL:     cfg.URL,
				User:    userName,
			}

			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, info)
			}

			cols := []output.Column{
				{Header: "PROFILE", Width: 15},
				{Header: "URL", Width: 40},
				{Header: "USER", Width: 25},
			}
			tf := output.NewTableFormatter(cols, flags.NoColor)
			rows := [][]string{{info.Profile, info.URL, info.User}}
			fmt.Print(tf.FormatRows(rows))
			return nil
		},
	}
}
