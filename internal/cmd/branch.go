package cmd

import (
	"context"
	"fmt"

	"github.com/manu/bb/internal/output"
	"github.com/spf13/cobra"
)

func NewBranchCmd(flags *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "branch",
		Short: "Manage branches",
	}
	cmd.AddCommand(newBranchListCmd(flags))
	cmd.AddCommand(newBranchCreateCmd(flags))
	cmd.AddCommand(newBranchDefaultCmd(flags))
	cmd.AddCommand(newBranchDeleteCmd(flags))
	return cmd
}

func newBranchListCmd(flags *GlobalFlags) *cobra.Command {
	var limit, page int
	var filter string

	cmd := &cobra.Command{
		Use:   "list [project] [repo]",
		Short: "List branches",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, _, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			start := (page - 1) * limit
			results, err := c.ListBranches(context.Background(), project, repo, filter, start, limit)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, results.Values)
			}
			cols := []output.Column{
				{Header: "NAME", Width: 40},
				{Header: "LATEST COMMIT", Width: 12},
				{Header: "DEFAULT", Width: 7},
			}
			tf := output.NewTableFormatter(cols, flags.NoColor)
			var rows [][]string
			for _, b := range results.Values {
				def := ""
				if b.IsDefault {
					def = "✓"
				}
				commit := b.LatestCommit
				if len(commit) > 12 {
					commit = commit[:12]
				}
				rows = append(rows, []string{b.DisplayID, commit, def})
			}
			fmt.Print(tf.FormatRows(rows))
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	cmd.Flags().StringVar(&filter, "filter", "", "filter branches by name")
	return cmd
}

func newBranchCreateCmd(flags *GlobalFlags) *cobra.Command {
	var from string

	cmd := &cobra.Command{
		Use:   "create [project] [repo] <name>",
		Short: "Create a branch",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, remaining, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			if len(remaining) < 1 {
				return fmt.Errorf("branch name is required")
			}
			name := remaining[0]
			if from == "" {
				return fmt.Errorf("--from is required (source branch or commit)")
			}
			branch, err := c.CreateBranch(context.Background(), project, repo, name, from)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, branch)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Branch '%s' created\n", branch.DisplayID)
			return nil
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "source branch or commit (required)")
	return cmd
}

func newBranchDefaultCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "default [project] [repo]",
		Short: "Get default branch",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, _, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			branch, err := c.GetDefaultBranch(context.Background(), project, repo)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, branch)
			}
			fmt.Println(branch.DisplayID)
			return nil
		},
	}
}

func newBranchDeleteCmd(flags *GlobalFlags) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete [project] [repo] <name>",
		Short: "Delete a branch [dangerous]",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, remaining, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			if len(remaining) < 1 {
				return fmt.Errorf("branch name is required")
			}
			name := remaining[0]
			if !ConfirmDangerous("branch", name, confirm) {
				return fmt.Errorf("deletion cancelled")
			}
			if err := c.DeleteBranch(context.Background(), project, repo, name); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Branch '%s' deleted\n", name)
			return nil
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "skip confirmation prompt")
	return cmd
}
