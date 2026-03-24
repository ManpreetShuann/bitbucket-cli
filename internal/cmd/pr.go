package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manu/bb/internal/client"
	"github.com/manu/bb/internal/output"
	"github.com/spf13/cobra"
)

func NewPRCmd(flags *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pr",
		Aliases: []string{"pull-request"},
		Short:   "Manage pull requests",
	}
	cmd.AddCommand(newPRListCmd(flags))
	cmd.AddCommand(newPRGetCmd(flags))
	cmd.AddCommand(newPRCreateCmd(flags))
	cmd.AddCommand(newPRUpdateCmd(flags))
	cmd.AddCommand(newPRMergeCmd(flags))
	cmd.AddCommand(newPRDeclineCmd(flags))
	cmd.AddCommand(newPRReopenCmd(flags))
	cmd.AddCommand(newPRApproveCmd(flags))
	cmd.AddCommand(newPRUnapproveCmd(flags))
	cmd.AddCommand(newPRRequestChangesCmd(flags))
	cmd.AddCommand(newPRRemoveRequestCmd(flags))
	cmd.AddCommand(newPRCanMergeCmd(flags))
	cmd.AddCommand(newPRDiffCmd(flags))
	cmd.AddCommand(newPRDiffStatCmd(flags))
	cmd.AddCommand(newPRCommitsCmd(flags))
	cmd.AddCommand(newPRActivitiesCmd(flags))
	cmd.AddCommand(newPRParticipantsCmd(flags))
	cmd.AddCommand(newPRWatchCmd(flags))
	cmd.AddCommand(newPRUnwatchCmd(flags))
	cmd.AddCommand(newPRDraftCmd(flags))
	cmd.AddCommand(newPRPublishCmd(flags))
	cmd.AddCommand(newPRConvertToDraftCmd(flags))
	cmd.AddCommand(newPRSuggestMessageCmd(flags))
	cmd.AddCommand(newPRDeleteCmd(flags))
	cmd.AddCommand(NewPRCommentCmd(flags))
	cmd.AddCommand(NewPRTaskCmd(flags))
	return cmd
}

func timeAgo(ts int64) string {
	if ts == 0 {
		return ""
	}
	d := time.Since(time.UnixMilli(ts))
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func reviewerStr(reviewers []client.Participant) string {
	var parts []string
	for _, r := range reviewers {
		indicator := "⏳"
		switch r.Status {
		case "APPROVED":
			indicator = "✓"
		case "NEEDS_WORK":
			indicator = "✗"
		}
		parts = append(parts, r.User.DisplayName+" "+indicator)
	}
	return strings.Join(parts, ", ")
}

func parsePRID(s string) (int, error) {
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid PR ID: %q", s)
	}
	return id, nil
}

func newPRListCmd(flags *GlobalFlags) *cobra.Command {
	var limit, page int
	var state, direction, order, author string
	var draft bool

	cmd := &cobra.Command{
		Use:   "list [project] [repo]",
		Short: "List pull requests",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, _, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			opts := client.ListPROptions{
				State:     state,
				Direction: direction,
				Order:     order,
				Start:     (page - 1) * limit,
				Limit:     limit,
			}
			if author != "" {
				opts.Participant = author
			}
			if cmd.Flags().Changed("draft") {
				opts.Draft = &draft
			}
			results, err := c.ListPullRequests(context.Background(), project, repo, opts)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, results.Values)
			}
			cols := []output.Column{
				{Header: "ID", Width: 6},
				{Header: "TITLE", Width: 40},
				{Header: "AUTHOR", Width: 15},
				{Header: "STATE", Width: 10},
				{Header: "REVIEWERS", Width: 30},
				{Header: "UPDATED", Width: 10},
			}
			tf := output.NewTableFormatter(cols, flags.NoColor)
			var rows [][]string
			for _, pr := range results.Values {
				rows = append(rows, []string{
					strconv.Itoa(pr.ID),
					pr.Title,
					pr.Author.User.DisplayName,
					pr.State,
					reviewerStr(pr.Reviewers),
					timeAgo(pr.UpdatedDate),
				})
			}
			fmt.Print(tf.FormatRows(rows))
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	cmd.Flags().StringVar(&state, "state", "", "filter by state (OPEN, DECLINED, MERGED, ALL)")
	cmd.Flags().StringVar(&direction, "direction", "", "INCOMING or OUTGOING")
	cmd.Flags().StringVar(&order, "order", "", "OLDEST or NEWEST")
	cmd.Flags().StringVar(&author, "author", "", "filter by author username")
	cmd.Flags().BoolVar(&draft, "draft", false, "filter draft PRs")
	return cmd
}

func newPRGetCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get [project] [repo] <id>",
		Short: "Get pull request details",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			pr, err := c.GetPullRequest(context.Background(), project, repo, id)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, pr)
			}
			fmt.Printf("ID:          %d\n", pr.ID)
			fmt.Printf("Title:       %s\n", pr.Title)
			fmt.Printf("State:       %s\n", pr.State)
			fmt.Printf("Draft:       %t\n", pr.Draft)
			fmt.Printf("Author:      %s\n", pr.Author.User.DisplayName)
			fmt.Printf("Source:      %s\n", pr.FromRef.DisplayID)
			fmt.Printf("Target:      %s\n", pr.ToRef.DisplayID)
			fmt.Printf("Description: %s\n", pr.Description)
			fmt.Printf("Reviewers:   %s\n", reviewerStr(pr.Reviewers))
			fmt.Printf("Updated:     %s\n", timeAgo(pr.UpdatedDate))
			return nil
		},
	}
}

func newPRCreateCmd(flags *GlobalFlags) *cobra.Command {
	var title, source, target, description string
	var reviewers []string
	var draft bool

	cmd := &cobra.Command{
		Use:   "create [project] [repo]",
		Short: "Create a pull request",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, _, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			if title == "" {
				return fmt.Errorf("--title is required")
			}
			if source == "" {
				return fmt.Errorf("--source is required")
			}
			if target == "" {
				target = "main"
			}
			pr, err := c.CreatePullRequest(context.Background(), project, repo, client.CreatePRInput{
				Title:       title,
				Description: description,
				FromRef:     source,
				ToRef:       target,
				Reviewers:   reviewers,
				Draft:       draft,
			})
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, pr)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Pull request #%d created\n", pr.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "PR title (required)")
	cmd.Flags().StringVar(&source, "source", "", "source branch (required)")
	cmd.Flags().StringVar(&target, "target", "", "target branch (default: main)")
	cmd.Flags().StringVar(&description, "description", "", "PR description")
	cmd.Flags().StringSliceVar(&reviewers, "reviewer", nil, "reviewer usernames (repeatable)")
	cmd.Flags().BoolVar(&draft, "draft", false, "create as draft")
	return cmd
}

func newPRUpdateCmd(flags *GlobalFlags) *cobra.Command {
	var title, description, target string
	var reviewers []string
	var draft bool
	var version int

	cmd := &cobra.Command{
		Use:   "update [project] [repo] <id>",
		Short: "Update a pull request",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			input := client.UpdatePRInput{Title: title, TargetBranch: target}
			if cmd.Flags().Changed("description") {
				input.Description = &description
			}
			if cmd.Flags().Changed("reviewer") {
				input.Reviewers = reviewers
			}
			if cmd.Flags().Changed("draft") {
				input.Draft = &draft
			}
			pr, err := c.UpdatePullRequest(context.Background(), project, repo, id, version, input)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, pr)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Pull request #%d updated\n", pr.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "new title")
	cmd.Flags().StringVar(&description, "description", "", "new description")
	cmd.Flags().StringVar(&target, "target", "", "new target branch")
	cmd.Flags().StringSliceVar(&reviewers, "reviewer", nil, "reviewer usernames")
	cmd.Flags().BoolVar(&draft, "draft", false, "set draft status")
	cmd.Flags().IntVar(&version, "version", 0, "PR version for optimistic locking")
	return cmd
}

func newPRMergeCmd(flags *GlobalFlags) *cobra.Command {
	var strategy string
	var version int

	cmd := &cobra.Command{
		Use:   "merge [project] [repo] <id>",
		Short: "Merge a pull request",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			pr, err := c.MergePullRequest(context.Background(), project, repo, id, version, strategy)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, pr)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Pull request #%d merged\n", pr.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&strategy, "strategy", "", "merge strategy")
	cmd.Flags().IntVar(&version, "version", 0, "PR version")
	return cmd
}

func newPRDeclineCmd(flags *GlobalFlags) *cobra.Command {
	var version int
	cmd := &cobra.Command{
		Use:   "decline [project] [repo] <id>",
		Short: "Decline a pull request",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			pr, err := c.DeclinePullRequest(context.Background(), project, repo, id, version)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, pr)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Pull request #%d declined\n", pr.ID)
			return nil
		},
	}
	cmd.Flags().IntVar(&version, "version", 0, "PR version")
	return cmd
}

func newPRReopenCmd(flags *GlobalFlags) *cobra.Command {
	var version int
	cmd := &cobra.Command{
		Use:   "reopen [project] [repo] <id>",
		Short: "Reopen a pull request",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			pr, err := c.ReopenPullRequest(context.Background(), project, repo, id, version)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, pr)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Pull request #%d reopened\n", pr.ID)
			return nil
		},
	}
	cmd.Flags().IntVar(&version, "version", 0, "PR version")
	return cmd
}

func newPRApproveCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "approve [project] [repo] <id>",
		Short: "Approve a pull request",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			_, err = c.ApprovePullRequest(context.Background(), project, repo, id)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Pull request #%d approved\n", id)
			return nil
		},
	}
}

func newPRUnapproveCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "unapprove [project] [repo] <id>",
		Short: "Remove approval from a pull request",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			if err := c.UnapprovePullRequest(context.Background(), project, repo, id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Approval removed from PR #%d\n", id)
			return nil
		},
	}
}

func newPRRequestChangesCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "request-changes [project] [repo] <id>",
		Short: "Request changes on a pull request",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			_, err = c.RequestChanges(context.Background(), project, repo, id)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Changes requested on PR #%d\n", id)
			return nil
		},
	}
}

func newPRRemoveRequestCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "remove-request [project] [repo] <id>",
		Short: "Remove change request from a pull request",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			_, err = c.RemoveChangeRequest(context.Background(), project, repo, id)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Change request removed from PR #%d\n", id)
			return nil
		},
	}
}

func newPRCanMergeCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "can-merge [project] [repo] <id>",
		Short: "Check if a pull request can be merged",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			status, err := c.CanMerge(context.Background(), project, repo, id)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, status)
			}
			if status.CanMerge {
				fmt.Println("✓ Can merge")
			} else {
				fmt.Println("✗ Cannot merge")
				if status.Conflicted {
					fmt.Println("  Reason: Conflicts detected")
				}
				for _, v := range status.Vetoes {
					fmt.Printf("  Veto: %s\n", v.SummaryMessage)
				}
			}
			return nil
		},
	}
}

func newPRDiffCmd(flags *GlobalFlags) *cobra.Command {
	var contextLines int
	cmd := &cobra.Command{
		Use:   "diff [project] [repo] <id>",
		Short: "Show pull request diff",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			diff, err := c.GetPullRequestDiff(context.Background(), project, repo, id, contextLines, "")
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
	return cmd
}

func newPRDiffStatCmd(flags *GlobalFlags) *cobra.Command {
	var limit, page int
	cmd := &cobra.Command{
		Use:   "diffstat [project] [repo] <id>",
		Short: "Show pull request diff statistics",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			start := (page - 1) * limit
			results, err := c.GetPullRequestDiffStat(context.Background(), project, repo, id, start, limit)
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

func newPRCommitsCmd(flags *GlobalFlags) *cobra.Command {
	var limit, page int
	cmd := &cobra.Command{
		Use:   "commits [project] [repo] <id>",
		Short: "List pull request commits",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			start := (page - 1) * limit
			results, err := c.ListPullRequestCommits(context.Background(), project, repo, id, start, limit)
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
				hash := cm.DisplayID
				msg := strings.Split(cm.Message, "\n")[0]
				rows = append(rows, []string{hash, cm.Author.Name, msg})
			}
			fmt.Print(tf.FormatRows(rows))
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	return cmd
}

func newPRActivitiesCmd(flags *GlobalFlags) *cobra.Command {
	var limit, page int
	cmd := &cobra.Command{
		Use:   "activities [project] [repo] <id>",
		Short: "List pull request activities",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			start := (page - 1) * limit
			results, err := c.GetPullRequestActivities(context.Background(), project, repo, id, start, limit)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, results.Values)
			}
			for _, a := range results.Values {
				fmt.Printf("[%s] %s - %s\n", timeAgo(a.CreatedDate), a.User.DisplayName, a.Action)
				if a.Comment != nil {
					fmt.Printf("  %s\n", a.Comment.Text)
				}
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	return cmd
}

func newPRParticipantsCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "participants [project] [repo] <id>",
		Short: "List pull request participants",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			results, err := c.ListPullRequestParticipants(context.Background(), project, repo, id, 0, 100)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, results.Values)
			}
			cols := []output.Column{
				{Header: "USER", Width: 20},
				{Header: "ROLE", Width: 15},
				{Header: "STATUS", Width: 12},
			}
			tf := output.NewTableFormatter(cols, flags.NoColor)
			var rows [][]string
			for _, p := range results.Values {
				rows = append(rows, []string{p.User.DisplayName, p.Role, p.Status})
			}
			fmt.Print(tf.FormatRows(rows))
			return nil
		},
	}
}

func newPRWatchCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "watch [project] [repo] <id>",
		Short: "Watch a pull request",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			if err := c.WatchPullRequest(context.Background(), project, repo, id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Watching PR #%d\n", id)
			return nil
		},
	}
}

func newPRUnwatchCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "unwatch [project] [repo] <id>",
		Short: "Unwatch a pull request",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			if err := c.UnwatchPullRequest(context.Background(), project, repo, id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Unwatched PR #%d\n", id)
			return nil
		},
	}
}

func newPRDraftCmd(flags *GlobalFlags) *cobra.Command {
	var title, source, target, description string
	var reviewers []string

	cmd := &cobra.Command{
		Use:   "draft [project] [repo]",
		Short: "Create a draft pull request",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, _, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			if title == "" {
				return fmt.Errorf("--title is required")
			}
			if source == "" {
				return fmt.Errorf("--source is required")
			}
			if target == "" {
				target = "main"
			}
			pr, err := c.CreatePullRequest(context.Background(), project, repo, client.CreatePRInput{
				Title:       title,
				Description: description,
				FromRef:     source,
				ToRef:       target,
				Reviewers:   reviewers,
				Draft:       true,
			})
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, pr)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Draft PR #%d created\n", pr.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "PR title (required)")
	cmd.Flags().StringVar(&source, "source", "", "source branch (required)")
	cmd.Flags().StringVar(&target, "target", "", "target branch")
	cmd.Flags().StringVar(&description, "description", "", "PR description")
	cmd.Flags().StringSliceVar(&reviewers, "reviewer", nil, "reviewer usernames")
	return cmd
}

func newPRPublishCmd(flags *GlobalFlags) *cobra.Command {
	var version int
	cmd := &cobra.Command{
		Use:   "publish [project] [repo] <id>",
		Short: "Publish a draft pull request",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			pr, err := c.PublishDraft(context.Background(), project, repo, id, version)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, pr)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ PR #%d published\n", pr.ID)
			return nil
		},
	}
	cmd.Flags().IntVar(&version, "version", 0, "PR version")
	return cmd
}

func newPRConvertToDraftCmd(flags *GlobalFlags) *cobra.Command {
	var version int
	cmd := &cobra.Command{
		Use:   "convert-to-draft [project] [repo] <id>",
		Short: "Convert a pull request to draft",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			pr, err := c.ConvertToDraft(context.Background(), project, repo, id, version)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, pr)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ PR #%d converted to draft\n", pr.ID)
			return nil
		},
	}
	cmd.Flags().IntVar(&version, "version", 0, "PR version")
	return cmd
}

func newPRSuggestMessageCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "suggest-message [project] [repo] <id>",
		Short: "Get commit message suggestion",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			suggestion, err := c.GetCommitMessageSuggestion(context.Background(), project, repo, id)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, suggestion)
			}
			if body, ok := suggestion["body"].(string); ok {
				fmt.Println(body)
			}
			return nil
		},
	}
}

func newPRDeleteCmd(flags *GlobalFlags) *cobra.Command {
	var confirm bool
	var version int

	cmd := &cobra.Command{
		Use:   "delete [project] [repo] <id>",
		Short: "Delete a pull request [dangerous]",
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
			id, err := parsePRID(remaining[0])
			if err != nil {
				return err
			}
			name := fmt.Sprintf("PR #%d", id)
			if !ConfirmDangerous("pull request", name, confirm) {
				return fmt.Errorf("deletion cancelled")
			}
			if err := c.DeletePullRequest(context.Background(), project, repo, id, version); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Pull request #%d deleted\n", id)
			return nil
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "skip confirmation prompt")
	cmd.Flags().IntVar(&version, "version", 0, "PR version")
	return cmd
}
