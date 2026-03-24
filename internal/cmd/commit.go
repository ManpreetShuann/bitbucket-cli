package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/manu/bb/internal/output"
	"github.com/spf13/cobra"
)

func NewCommitCmd(flags *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Manage commits",
	}
	cmd.AddCommand(newCommitListCmd(flags))
	cmd.AddCommand(newCommitGetCmd(flags))
	cmd.AddCommand(newCommitDiffCmd(flags))
	cmd.AddCommand(newCommitChangesCmd(flags))
	return cmd
}

func newCommitListCmd(flags *GlobalFlags) *cobra.Command {
	var limit, page int
	var until, since, path string

	cmd := &cobra.Command{
		Use:   "list [project] [repo]",
		Short: "List commits",
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
			results, err := c.ListCommits(context.Background(), project, repo, until, since, path, start, limit)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, results.Values)
			}
			cols := []output.Column{
				{Header: "COMMIT", Width: 12},
				{Header: "AUTHOR", Width: 20},
				{Header: "MESSAGE", Width: 60},
			}
			tf := output.NewTableFormatter(cols, flags.NoColor)
			var rows [][]string
			for _, cm := range results.Values {
				msg := strings.Split(cm.Message, "\n")[0]
				rows = append(rows, []string{cm.DisplayID, cm.Author.Name, msg})
			}
			fmt.Print(tf.FormatRows(rows))
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	cmd.Flags().StringVar(&until, "until", "", "commit hash or ref upper bound")
	cmd.Flags().StringVar(&since, "since", "", "commit hash or ref lower bound")
	cmd.Flags().StringVar(&path, "path", "", "filter by file path")
	return cmd
}

func newCommitGetCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get [project] [repo] <commit-id>",
		Short: "Get commit details",
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
				return fmt.Errorf("commit ID is required")
			}
			commit, err := c.GetCommit(context.Background(), project, repo, remaining[0])
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, commit)
			}
			fmt.Printf("Commit:  %s\n", commit.ID)
			fmt.Printf("Author:  %s <%s>\n", commit.Author.Name, commit.Author.EmailAddress)
			fmt.Printf("Message: %s\n", commit.Message)
			return nil
		},
	}
}

func newCommitDiffCmd(flags *GlobalFlags) *cobra.Command {
	var contextLines int
	var srcPath string

	cmd := &cobra.Command{
		Use:   "diff [project] [repo] <commit-id>",
		Short: "Show commit diff",
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
				return fmt.Errorf("commit ID is required")
			}
			diff, err := c.GetCommitDiff(context.Background(), project, repo, remaining[0], contextLines, srcPath)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, diff)
			}
			for _, d := range diff.Diffs {
				if d.Source != nil {
					fmt.Printf("--- a/%s\n", d.Source.ToString)
				}
				if d.Destination != nil {
					fmt.Printf("+++ b/%s\n", d.Destination.ToString)
				}
				for _, h := range d.Hunks {
					fmt.Printf("@@ -%d,%d +%d,%d @@\n", h.SourceLine, h.SourceSpan, h.DestinationLine, h.DestinationSpan)
					for _, s := range h.Segments {
						prefix := " "
						if s.Type == "ADDED" {
							prefix = "+"
						} else if s.Type == "REMOVED" {
							prefix = "-"
						}
						for _, l := range s.Lines {
							fmt.Printf("%s%s\n", prefix, l.Line)
						}
					}
				}
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&contextLines, "context", 10, "context lines")
	cmd.Flags().StringVar(&srcPath, "src-path", "", "source file path")
	return cmd
}

func newCommitChangesCmd(flags *GlobalFlags) *cobra.Command {
	var limit, page int
	cmd := &cobra.Command{
		Use:   "changes [project] [repo] <commit-id>",
		Short: "List files changed in a commit",
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
				return fmt.Errorf("commit ID is required")
			}
			start := (page - 1) * limit
			results, err := c.GetCommitChanges(context.Background(), project, repo, remaining[0], start, limit)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, results.Values)
			}
			cols := []output.Column{
				{Header: "TYPE", Width: 10},
				{Header: "PATH", Width: 60},
			}
			tf := output.NewTableFormatter(cols, flags.NoColor)
			var rows [][]string
			for _, ch := range results.Values {
				rows = append(rows, []string{ch.Type, ch.Path.ToString})
			}
			fmt.Print(tf.FormatRows(rows))
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	return cmd
}
