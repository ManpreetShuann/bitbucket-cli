package cmd

import (
	"context"
	"fmt"

	"github.com/manu/bb/internal/client"
	"github.com/manu/bb/internal/config"
	"github.com/manu/bb/internal/output"
	"github.com/manu/bb/internal/validation"
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

func repoColumns() []output.Column {
	return []output.Column{
		{Header: "SLUG", Width: 20},
		{Header: "NAME", Width: 25},
		{Header: "DESCRIPTION", Width: 35},
		{Header: "STATE", Width: 12},
		{Header: "PROJECT", Width: 10},
	}
}

func repoRow(r client.Repository) []string {
	return []string{r.Slug, r.Name, r.Description, r.State, r.Project.Key}
}

func newRepoListCmd(flags *GlobalFlags) *cobra.Command {
	var limitFlag, page int
	var allPages bool

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
			limit := validation.ClampLimit(limitFlag)
			start := (page - 1) * limit

			var repos []client.Repository
			if allPages {
				repos, err = client.GetAll[client.Repository](ctx, c, "/projects/"+project+"/repos", nil, limit)
				if err != nil {
					return err
				}
			} else {
				resp, err := c.ListRepositories(ctx, project, start, limit)
				if err != nil {
					return err
				}
				repos = resp.Values
			}

			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, repos)
			}

			tf := output.NewTableFormatter(repoColumns(), flags.NoColor)
			var rows [][]string
			for _, r := range repos {
				rows = append(rows, repoRow(r))
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

			tf := output.NewTableFormatter(repoColumns(), flags.NoColor)
			rows := [][]string{repoRow(*r)}
			fmt.Print(tf.FormatRows(rows))
			return nil
		},
	}
}

func newRepoCreateCmd(flags *GlobalFlags) *cobra.Command {
	var description string
	var forkable bool

	cmd := &cobra.Command{
		Use:   "create <project> <name>",
		Short: "Create a repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("project and name are required")
			}
			project := args[0]
			name := args[1]

			c, _, err := newClient(flags)
			if err != nil {
				return err
			}
			r, err := c.CreateRepository(context.Background(), project, name, description, forkable)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, r)
			}

			tf := output.NewTableFormatter(repoColumns(), flags.NoColor)
			rows := [][]string{repoRow(*r)}
			fmt.Print(tf.FormatRows(rows))
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
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, _, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			if !ConfirmDestructive("repository", repo, confirm, understand) {
				return fmt.Errorf("deletion cancelled")
			}
			if err := c.DeleteRepository(context.Background(), project, repo); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "Repository '%s' deleted\n", repo)
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
			project := args[0]
			repo := args[1]

			cfg, err := config.Load(flags.Profile, "")
			if err != nil {
				return err
			}

			configDir := config.ConfigDir()
			if err := config.SaveProfile(configDir, flags.Profile, cfg.URL, project, repo); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "Default context set to %s/%s\n", project, repo)
			return nil
		},
	}
}

func newRepoClearCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear default project and repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir := config.ConfigDir()
			if err := config.ClearDefaults(configDir, flags.Profile); err != nil {
				return err
			}
			fmt.Fprintln(cmd.ErrOrStderr(), "Default context cleared")
			return nil
		},
	}
}
