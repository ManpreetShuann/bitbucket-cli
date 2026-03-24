package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ManpreetShuann/bitbucket-cli/internal/client"
	"github.com/ManpreetShuann/bitbucket-cli/internal/output"
	"github.com/ManpreetShuann/bitbucket-cli/internal/validation"
	"github.com/spf13/cobra"
)

func NewProjectCmd(flags *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects",
	}
	cmd.AddCommand(newProjectListCmd(flags))
	cmd.AddCommand(newProjectGetCmd(flags))
	cmd.AddCommand(newProjectDeleteCmd(flags))
	return cmd
}

func projectColumns() []output.Column {
	return []output.Column{
		{Header: "KEY", Width: 10},
		{Header: "NAME", Width: 30},
		{Header: "DESCRIPTION", Width: 40},
		{Header: "PUBLIC", Width: 8},
		{Header: "TYPE", Width: 15},
	}
}

func projectRow(p client.Project) []string {
	return []string{p.Key, p.Name, p.Description, strconv.FormatBool(p.Public), p.Type}
}

func newProjectListCmd(flags *GlobalFlags) *cobra.Command {
	var limitFlag, page int
	var allPages bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, _, err := newClient(flags)
			if err != nil {
				return err
			}
			ctx := context.Background()
			limit := validation.ClampLimit(limitFlag)
			start := (page - 1) * limit

			var projects []client.Project
			if allPages {
				projects, err = client.GetAll[client.Project](ctx, c, "/projects", nil, limit)
				if err != nil {
					return err
				}
			} else {
				resp, err := c.ListProjects(ctx, start, limit)
				if err != nil {
					return err
				}
				projects = resp.Values
			}

			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, projects)
			}

			tf := output.NewTableFormatter(projectColumns(), flags.NoColor)
			var rows [][]string
			for _, p := range projects {
				rows = append(rows, projectRow(p))
			}
			fmt.Print(tf.FormatRows(rows))
			return nil
		},
	}
	cmd.Flags().IntVar(&limitFlag, "limit", 25, "items per page")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	cmd.Flags().BoolVar(&allPages, "all", false, "fetch all results")
	return cmd
}

func newProjectGetCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get project details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			if err := validation.ValidateProjectKey(key); err != nil {
				return err
			}
			c, _, err := newClient(flags)
			if err != nil {
				return err
			}
			project, err := c.GetProject(context.Background(), key)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, project)
			}

			tf := output.NewTableFormatter(projectColumns(), flags.NoColor)
			rows := [][]string{projectRow(*project)}
			fmt.Print(tf.FormatRows(rows))
			return nil
		},
	}
}

func newProjectDeleteCmd(flags *GlobalFlags) *cobra.Command {
	var confirm, understand bool

	cmd := &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a project [destructive]",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			if !ConfirmDestructive("project", key, confirm, understand) {
				return fmt.Errorf("deletion cancelled")
			}
			c, _, err := newClient(flags)
			if err != nil {
				return err
			}
			if err := c.DeleteProject(context.Background(), key); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "Project '%s' deleted\n", key)
			return nil
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "skip confirmation prompt")
	cmd.Flags().BoolVar(&understand, "i-understand-this-is-destructive", false, "required with --confirm for scripting")
	return cmd
}
