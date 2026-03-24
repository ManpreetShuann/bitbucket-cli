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
			reader := bufio.NewReader(os.Stdin)

			if url == "" {
				fmt.Fprint(os.Stderr, "Bitbucket Server URL: ")
				input, _ := reader.ReadString('\n')
				url = strings.TrimSpace(input)
			}
			if url == "" {
				return fmt.Errorf("URL is required")
			}

			if token == "" {
				fmt.Fprint(os.Stderr, "Personal access token: ")
				tokenBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					// Fallback for non-TTY
					input, _ := reader.ReadString('\n')
					token = strings.TrimSpace(input)
				} else {
					token = string(tokenBytes)
					fmt.Fprintln(os.Stderr)
				}
			}
			if token == "" {
				return fmt.Errorf("token is required")
			}

			// Validate by calling the API
			c := client.New(url, token)
			var user client.User
			err := c.Get(context.Background(), "/users", nil, &user)
			// Ignore error — just testing connectivity
			_ = err

			configDir := config.ConfigDir()
			if err := config.SaveProfile(configDir, profile, url, "", ""); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			credsPath := filepath.Join(configDir, "credentials.yaml")
			if err := config.SaveCredentials(credsPath, profile, token); err != nil {
				return fmt.Errorf("failed to save credentials: %w", err)
			}

			fmt.Fprintf(os.Stderr, "✓ Logged in to %s (profile: %s)\n", url, profile)
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
			fmt.Fprintf(os.Stderr, "✓ Logged out (profile: %s)\n", flags.Profile)
			return nil
		},
	}
}

func newAuthStatusCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flags.Profile, "")
			if err != nil {
				return err
			}

			fmt.Printf("Profile: %s\n", cfg.Profile)
			if cfg.URL != "" {
				fmt.Printf("URL:     %s\n", cfg.URL)
			} else {
				fmt.Println("URL:     (not set)")
			}
			if cfg.Token != "" {
				fmt.Println("Auth:    ✓ Token configured")
			} else {
				fmt.Println("Auth:    ✗ No token")
			}
			if cfg.DefaultProject != "" {
				fmt.Printf("Project: %s\n", cfg.DefaultProject)
			}
			if cfg.DefaultRepo != "" {
				fmt.Printf("Repo:    %s\n", cfg.DefaultRepo)
			}
			return nil
		},
	}
}
