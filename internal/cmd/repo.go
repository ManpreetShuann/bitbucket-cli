package cmd

import (
	"context"
	"fmt"

	"github.com/manu/bb/internal/config"
	"github.com/manu/bb/internal/output"
	"github.com/spf13/cobra"
)

func NewRepoCmd(flags *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage repositories",
	}
	cmd.AddCommand(newRepoListCmd(flags))
	cmd.AddCommand(newRepoGetCmd(flags))
	cmd.AddCommand(newRepoCreateCmd(flags))
	cmd.AddCommand(newRepoDeleteCmd(flags))
	cmd.AddCommand(newRepoUseCmd(flags))
	cmd.AddCommand(newRepoClearCmd(flags))
	return cmd
}

func newRepoListCmd(flags *GlobalFlags) *cobra.Command {
	var limit, page int
	cmd := &cobra.Command{
		Use:   "list [project]",
		Short: "List repositories in a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, _, _, err := resolveProjectRepo(cfg, args, 1)
			if err != nil {
				return err
			}
			ctx := context.Background()
			start := (page - 1) * limit
			results, err := c.ListRepositories(ctx, project, start, limit)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, results.Values)
			}
			cols := []output.Column{
				{Header: "SLUG", Width: 25},
				{Header: "NAME", Width: 30},
				{Header: "STATE", Width: 10},
				{Header: "DESCRIPTION", Width: 40},
			}
			tf := output.NewTableFormatter(cols, flags.NoColor)
			var rows [][]string
			for _, r := range results.Values {
				rows = append(rows, []string{r.Slug, r.Name, r.State, r.Description})
			}
			fmt.Print(tf.FormatRows(rows))
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	return cmd
}

func newRepoGetCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get [project] [repo]",
		Short: "Get repository details",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, _, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			r, err := c.GetRepository(context.Background(), project, repo)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, r)
			}
			fmt.Printf("Slug:        %s\n", r.Slug)
			fmt.Printf("Name:        %s\n", r.Name)
			fmt.Printf("Project:     %s\n", r.Project.Key)
			fmt.Printf("State:       %s\n", r.State)
			fmt.Printf("Forkable:    %t\n", r.Forkable)
			fmt.Printf("Description: %s\n", r.Description)
			return nil
		},
	}
}

func newRepoCreateCmd(flags *GlobalFlags) *cobra.Command {
	var description string
	var forkable bool

	cmd := &cobra.Command{
		Use:   "create [project] <name>",
		Short: "Create a repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			var project, name string
			if len(args) >= 2 {
				project = args[0]
				name = args[1]
			} else if len(args) == 1 && cfg.DefaultProject != "" {
				project = cfg.DefaultProject
				name = args[0]
			} else {
				return fmt.Errorf("project and name are required")
			}

			r, err := c.CreateRepository(context.Background(), project, name, description, forkable)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, r)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Repository '%s/%s' created\n", project, r.Slug)
			return nil
		},
	}
	cmd.Flags().StringVar(&description, "description", "", "repository description")
	cmd.Flags().BoolVar(&forkable, "forkable", true, "allow forking")
	return cmd
}

func newRepoDeleteCmd(flags *GlobalFlags) *cobra.Command {
	var confirm, understand bool

	cmd := &cobra.Command{
		Use:   "delete [project] [repo]",
		Short: "Delete a repository [destructive]",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, _, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			name := project + "/" + repo
			if !ConfirmDestructive("repository", name, confirm, understand) {
				return fmt.Errorf("deletion cancelled")
			}
			c, _, err := newClient(flags)
			if err != nil {
				return err
			}
			if err := c.DeleteRepository(context.Background(), project, repo); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Repository '%s' deleted\n", name)
			return nil
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "skip confirmation prompt")
	cmd.Flags().BoolVar(&understand, "i-understand-this-is-destructive", false, "required with --confirm for scripting")
	return cmd
}

func newRepoUseCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "use <project> <repo>",
		Short: "Set default project and repository",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir := config.ConfigDir()
			if err := config.SaveProfile(configDir, flags.Profile, "", args[0], args[1]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Default set to %s/%s\n", args[0], args[1])
			return nil
		},
	}
}

func newRepoClearCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear default project and repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.ClearDefaults(config.ConfigDir(), flags.Profile); err != nil {
				return err
			}
			fmt.Fprintln(cmd.ErrOrStderr(), "✓ Defaults cleared")
			return nil
		},
	}
}
