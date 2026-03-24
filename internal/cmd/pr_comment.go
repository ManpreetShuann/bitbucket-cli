package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manu/bb/internal/client"
	"github.com/manu/bb/internal/output"
	"github.com/spf13/cobra"
)

func NewPRCommentCmd(flags *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Manage PR comments",
	}
	cmd.AddCommand(newPRCommentListCmd(flags))
	cmd.AddCommand(newPRCommentGetCmd(flags))
	cmd.AddCommand(newPRCommentAddCmd(flags))
	cmd.AddCommand(newPRCommentUpdateCmd(flags))
	cmd.AddCommand(newPRCommentResolveCmd(flags))
	cmd.AddCommand(newPRCommentReopenCmd(flags))
	cmd.AddCommand(newPRCommentDeleteCmd(flags))
	return cmd
}

func newPRCommentListCmd(flags *GlobalFlags) *cobra.Command {
	var limit, page int
	cmd := &cobra.Command{
		Use:   "list [project] [repo] <pr-id>",
		Short: "List PR comments",
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
				return fmt.Errorf("PR ID is required")
			}
			prID, err := strconv.Atoi(remaining[0])
			if err != nil {
				return fmt.Errorf("invalid PR ID: %q", remaining[0])
			}
			start := (page - 1) * limit
			results, err := c.ListPRComments(context.Background(), project, repo, prID, start, limit)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, results.Values)
			}
			cols := []output.Column{
				{Header: "ID", Width: 8},
				{Header: "AUTHOR", Width: 15},
				{Header: "SEVERITY", Width: 10},
				{Header: "STATE", Width: 10},
				{Header: "TEXT", Width: 50},
			}
			tf := output.NewTableFormatter(cols, flags.NoColor)
			var rows [][]string
			for _, cm := range results.Values {
				text := cm.Text
				if len(text) > 50 {
					text = text[:47] + "..."
				}
				rows = append(rows, []string{strconv.Itoa(cm.ID), cm.Author.DisplayName, cm.Severity, cm.State, text})
			}
			fmt.Print(tf.FormatRows(rows))
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	return cmd
}

func newPRCommentGetCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get [project] [repo] <pr-id> <comment-id>",
		Short: "Get a PR comment",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, remaining, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			if len(remaining) < 2 {
				return fmt.Errorf("PR ID and comment ID are required")
			}
			prID, _ := strconv.Atoi(remaining[0])
			commentID, _ := strconv.Atoi(remaining[1])
			comment, err := c.GetPRComment(context.Background(), project, repo, prID, commentID)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, comment)
			}
			fmt.Printf("ID:       %d\n", comment.ID)
			fmt.Printf("Author:   %s\n", comment.Author.DisplayName)
			fmt.Printf("Severity: %s\n", comment.Severity)
			fmt.Printf("State:    %s\n", comment.State)
			fmt.Printf("Text:     %s\n", comment.Text)
			return nil
		},
	}
}

func newPRCommentAddCmd(flags *GlobalFlags) *cobra.Command {
	var text, file, lineType, fileType, severity string
	var line, replyTo int
	var blocker bool

	cmd := &cobra.Command{
		Use:   "add [project] [repo] <pr-id>",
		Short: "Add a comment to a PR",
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
				return fmt.Errorf("PR ID is required")
			}
			prID, _ := strconv.Atoi(remaining[0])
			if text == "" {
				return fmt.Errorf("--text is required")
			}
			input := client.AddCommentInput{Text: text, FilePath: file, LineType: lineType, FileType: fileType}
			if severity != "" {
				input.Severity = severity
			}
			if blocker {
				input.Severity = "BLOCKER"
			}
			if cmd.Flags().Changed("line") {
				input.Line = &line
			}
			if cmd.Flags().Changed("reply-to") {
				input.ParentID = &replyTo
			}
			comment, err := c.AddPRComment(context.Background(), project, repo, prID, input)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, comment)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Comment #%d added\n", comment.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&text, "text", "", "comment text (required)")
	cmd.Flags().StringVar(&file, "file", "", "file path for inline comment")
	cmd.Flags().IntVar(&line, "line", 0, "line number for inline comment")
	cmd.Flags().StringVar(&lineType, "line-type", "", "line type (ADDED, REMOVED, CONTEXT)")
	cmd.Flags().StringVar(&fileType, "file-type", "", "file type (FROM, TO)")
	cmd.Flags().IntVar(&replyTo, "reply-to", 0, "parent comment ID for replies")
	cmd.Flags().BoolVar(&blocker, "blocker", false, "mark as blocker")
	cmd.Flags().StringVar(&severity, "severity", "", "comment severity (NORMAL, BLOCKER)")
	return cmd
}

func newPRCommentUpdateCmd(flags *GlobalFlags) *cobra.Command {
	var text string
	var version int

	cmd := &cobra.Command{
		Use:   "update [project] [repo] <pr-id> <comment-id>",
		Short: "Update a PR comment",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, remaining, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			if len(remaining) < 2 {
				return fmt.Errorf("PR ID and comment ID are required")
			}
			prID, _ := strconv.Atoi(remaining[0])
			commentID, _ := strconv.Atoi(remaining[1])
			comment, err := c.UpdatePRComment(context.Background(), project, repo, prID, commentID, version, text)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, comment)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Comment #%d updated\n", comment.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&text, "text", "", "new comment text (required)")
	cmd.Flags().IntVar(&version, "version", 0, "comment version")
	return cmd
}

func newPRCommentResolveCmd(flags *GlobalFlags) *cobra.Command {
	var version int
	cmd := &cobra.Command{
		Use:   "resolve [project] [repo] <pr-id> <comment-id>",
		Short: "Resolve a PR comment",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, remaining, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			if len(remaining) < 2 {
				return fmt.Errorf("PR ID and comment ID are required")
			}
			prID, _ := strconv.Atoi(remaining[0])
			commentID, _ := strconv.Atoi(remaining[1])
			_, err = c.ResolvePRComment(context.Background(), project, repo, prID, commentID, version)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Comment #%d resolved\n", commentID)
			return nil
		},
	}
	cmd.Flags().IntVar(&version, "version", 0, "comment version")
	return cmd
}

func newPRCommentReopenCmd(flags *GlobalFlags) *cobra.Command {
	var version int
	cmd := &cobra.Command{
		Use:   "reopen [project] [repo] <pr-id> <comment-id>",
		Short: "Reopen a resolved PR comment",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, remaining, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			if len(remaining) < 2 {
				return fmt.Errorf("PR ID and comment ID are required")
			}
			prID, _ := strconv.Atoi(remaining[0])
			commentID, _ := strconv.Atoi(remaining[1])
			_, err = c.ReopenPRComment(context.Background(), project, repo, prID, commentID, version)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Comment #%d reopened\n", commentID)
			return nil
		},
	}
	cmd.Flags().IntVar(&version, "version", 0, "comment version")
	return cmd
}

func newPRCommentDeleteCmd(flags *GlobalFlags) *cobra.Command {
	var confirm bool
	var version int

	cmd := &cobra.Command{
		Use:   "delete [project] [repo] <pr-id> <comment-id>",
		Short: "Delete a PR comment [dangerous]",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, remaining, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			if len(remaining) < 2 {
				return fmt.Errorf("PR ID and comment ID are required")
			}
			prID, _ := strconv.Atoi(remaining[0])
			commentID, _ := strconv.Atoi(remaining[1])
			name := fmt.Sprintf("comment #%d", commentID)
			if !ConfirmDangerous("comment", name, confirm) {
				return fmt.Errorf("deletion cancelled")
			}
			if err := c.DeletePRComment(context.Background(), project, repo, prID, commentID, version); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Comment #%d deleted\n", commentID)
			return nil
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "skip confirmation prompt")
	cmd.Flags().IntVar(&version, "version", 0, "comment version")
	return cmd
}
