package cmd

import (
	"context"
	"fmt"

	"github.com/manu/bb/internal/output"
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

func newProjectListCmd(flags *GlobalFlags) *cobra.Command {
	var limit, page int
	var all bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, _, err := newClient(flags)
			if err != nil {
				return err
			}
			ctx := context.Background()
			start := (page - 1) * limit

			if flags.JSON || flags.Format != "" {
				if all {
					results, err := c.ListProjects(ctx, 0, limit)
					if err != nil {
						return err
					}
					return printFormatted(flags, results.Values)
				}
				results, err := c.ListProjects(ctx, start, limit)
				if err != nil {
					return err
				}
				return printFormatted(flags, results.Values)
			}

			results, err := c.ListProjects(ctx, start, limit)
			if err != nil {
				return err
			}

			cols := []output.Column{
				{Header: "KEY", Width: 15},
				{Header: "NAME", Width: 30},
				{Header: "DESCRIPTION", Width: 50},
			}
			tf := output.NewTableFormatter(cols, flags.NoColor)
			var rows [][]string
			for _, p := range results.Values {
				rows = append(rows, []string{p.Key, p.Name, p.Description})
			}
			fmt.Print(tf.FormatRows(rows))
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	cmd.Flags().BoolVar(&all, "all", false, "fetch all results")
	return cmd
}

func newProjectGetCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get project details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, _, err := newClient(flags)
			if err != nil {
				return err
			}
			project, err := c.GetProject(context.Background(), args[0])
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, project)
			}
			fmt.Printf("Key:         %s\n", project.Key)
			fmt.Printf("Name:        %s\n", project.Name)
			fmt.Printf("Description: %s\n", project.Description)
			fmt.Printf("Public:      %t\n", project.Public)
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
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Project '%s' deleted\n", key)
			return nil
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "skip confirmation prompt")
	cmd.Flags().BoolVar(&understand, "i-understand-this-is-destructive", false, "required with --confirm for scripting")
	return cmd
}
