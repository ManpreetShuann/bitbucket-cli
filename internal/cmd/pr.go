package cmd

import (
"context"
"fmt"
"strconv"
"strings"
"time"

"github.com/manu/bb/internal/client"
"github.com/manu/bb/internal/output"
"github.com/manu/bb/internal/validation"
"github.com/spf13/cobra"
)

func NewPRCmd(flags *GlobalFlags) *cobra.Command {
cmd := &cobra.Command{
Use:     "pr",
Aliases: []string{"pull-request"},
Short:   "Pull request commands",
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
cmd.AddCommand(newPRPublishCmd(flags))
cmd.AddCommand(newPRConvertToDraftCmd(flags))
cmd.AddCommand(newPRDraftCmd(flags))
cmd.AddCommand(newPRSuggestMessageCmd(flags))
cmd.AddCommand(newPRDeleteCmd(flags))
cmd.AddCommand(NewPRCommentCmd(flags))
cmd.AddCommand(NewPRTaskCmd(flags))
return cmd
}

func parsePRID(remaining []string) (int, error) {
if len(remaining) == 0 {
return 0, fmt.Errorf("pull request ID is required")
}
id, err := strconv.Atoi(remaining[0])
if err != nil || id <= 0 {
return 0, fmt.Errorf("invalid pull request ID: %s", remaining[0])
}
return id, nil
}


func formatReviewers(reviewers []client.Participant) string {
var parts []string
for _, r := range reviewers {
indicator := "⏳"
if r.Status == "APPROVED" {
indicator = "✓"
}
if r.Status == "NEEDS_WORK" {
indicator = "✗"
}
parts = append(parts, indicator+r.User.DisplayName)
}
return strings.Join(parts, ", ")
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
case d < 30*24*time.Hour:
return fmt.Sprintf("%dd ago", int(d.Hours()/24))
default:
return time.UnixMilli(ts).Format("2006-01-02")
}
}

func newPRListCmd(flags *GlobalFlags) *cobra.Command {
var limit, page int
var state, direction, order, author, draft string
var all bool

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
ctx := context.Background()

state = strings.ToUpper(state)
if err := validation.ValidatePRState(state); err != nil {
return err
}
if direction != "" {
direction = strings.ToUpper(direction)
if err := validation.ValidatePRDirection(direction); err != nil {
return err
}
}
if order != "" {
order = strings.ToUpper(order)
if err := validation.ValidatePROrder(order); err != nil {
return err
}
}

opts := client.ListPROptions{
State:     state,
Direction: direction,
Order:     order,
Limit:     limit,
}
if author != "" {
opts.Participant = author
}
if cmd.Flags().Changed("draft") {
d := strings.ToLower(draft) == "true"
opts.Draft = &d
}

var prs []client.PullRequest
if all {
opts.Start = 0
for {
results, err := c.ListPullRequests(ctx, project, repo, opts)
if err != nil {
return err
}
prs = append(prs, results.Values...)
if results.IsLastPage {
break
}
opts.Start = results.NextPageStart
}
} else {
opts.Start = (page - 1) * limit
results, err := c.ListPullRequests(ctx, project, repo, opts)
if err != nil {
return err
}
prs = results.Values
}

if flags.JSON || flags.Format != "" {
return printFormatted(flags, prs)
}

cols := []output.Column{
{Header: "ID", Width: 8},
{Header: "TITLE", Width: 35},
{Header: "AUTHOR", Width: 15},
{Header: "STATE", Width: 10},
{Header: "REVIEWERS", Width: 25},
{Header: "UPDATED", Width: 12},
}
tf := output.NewTableFormatter(cols, flags.NoColor)
var rows [][]string
for _, pr := range prs {
rows = append(rows, []string{
strconv.Itoa(pr.ID),
pr.Title,
pr.Author.User.DisplayName,
pr.State,
formatReviewers(pr.Reviewers),
timeAgo(pr.UpdatedDate),
})
}
fmt.Print(tf.FormatRows(rows))
return nil
},
}
cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
cmd.Flags().IntVar(&page, "page", 1, "page number")
cmd.Flags().StringVar(&state, "state", "OPEN", "filter by state (OPEN, DECLINED, MERGED, ALL)")
cmd.Flags().StringVar(&direction, "direction", "", "INCOMING or OUTGOING")
cmd.Flags().StringVar(&order, "order", "", "OLDEST or NEWEST")
cmd.Flags().StringVar(&author, "author", "", "filter by author username")
cmd.Flags().StringVar(&draft, "draft", "", "filter draft PRs (true/false)")
cmd.Flags().BoolVar(&all, "all", false, "fetch all results")
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
id, err := parsePRID(remaining)
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
fmt.Printf("Reviewers:   %s\n", formatReviewers(pr.Reviewers))
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
return fmt.Errorf("--target is required")
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
cmd.Flags().StringVar(&target, "target", "", "target branch (required)")
cmd.Flags().StringVar(&description, "description", "", "PR description")
cmd.Flags().StringSliceVar(&reviewers, "reviewer", nil, "reviewer usernames (repeatable)")
cmd.Flags().BoolVar(&draft, "draft", false, "create as draft")
return cmd
}

func newPRUpdateCmd(flags *GlobalFlags) *cobra.Command {
var title, description, target, draft string
var reviewers []string

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
id, err := parsePRID(remaining)
if err != nil {
return err
}
ctx := context.Background()
current, err := c.GetPullRequest(ctx, project, repo, id)
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
d := strings.ToLower(draft) == "true"
input.Draft = &d
}
pr, err := c.UpdatePullRequest(ctx, project, repo, id, current.Version, input)
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
cmd.Flags().StringVar(&draft, "draft", "", "set draft status (true/false)")
return cmd
}

func newPRMergeCmd(flags *GlobalFlags) *cobra.Command {
var strategy string

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
id, err := parsePRID(remaining)
if err != nil {
return err
}
ctx := context.Background()
current, err := c.GetPullRequest(ctx, project, repo, id)
if err != nil {
return err
}
pr, err := c.MergePullRequest(ctx, project, repo, id, current.Version, strategy)
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
return cmd
}

func newPRDeclineCmd(flags *GlobalFlags) *cobra.Command {
return &cobra.Command{
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
id, err := parsePRID(remaining)
if err != nil {
return err
}
ctx := context.Background()
current, err := c.GetPullRequest(ctx, project, repo, id)
if err != nil {
return err
}
pr, err := c.DeclinePullRequest(ctx, project, repo, id, current.Version)
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
}

func newPRReopenCmd(flags *GlobalFlags) *cobra.Command {
return &cobra.Command{
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
id, err := parsePRID(remaining)
if err != nil {
return err
}
ctx := context.Background()
current, err := c.GetPullRequest(ctx, project, repo, id)
if err != nil {
return err
}
pr, err := c.ReopenPullRequest(ctx, project, repo, id, current.Version)
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
id, err := parsePRID(remaining)
if err != nil {
return err
}
_, err = c.ApprovePullRequest(context.Background(), project, repo, id)
if err != nil {
return err
}
fmt.Fprintf(cmd.ErrOrStderr(), "✓ PR #%d approved\n", id)
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
id, err := parsePRID(remaining)
if err != nil {
return err
}
if err := c.UnapprovePullRequest(context.Background(), project, repo, id); err != nil {
return err
}
fmt.Fprintf(cmd.ErrOrStderr(), "Approval removed for PR #%d\n", id)
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
id, err := parsePRID(remaining)
if err != nil {
return err
}
_, err = c.RequestChanges(context.Background(), project, repo, id)
if err != nil {
return err
}
fmt.Fprintf(cmd.ErrOrStderr(), "Changes requested on PR #%d\n", id)
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
id, err := parsePRID(remaining)
if err != nil {
return err
}
_, err = c.RemoveChangeRequest(context.Background(), project, repo, id)
if err != nil {
return err
}
fmt.Fprintf(cmd.ErrOrStderr(), "Change request removed from PR #%d\n", id)
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
id, err := parsePRID(remaining)
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
cols := []output.Column{
{Header: "CAN_MERGE", Width: 12},
{Header: "CONFLICTED", Width: 12},
{Header: "VETOES", Width: 40},
}
tf := output.NewTableFormatter(cols, flags.NoColor)
vetoCount := strconv.Itoa(len(status.Vetoes))
rows := [][]string{
{strconv.FormatBool(status.CanMerge), strconv.FormatBool(status.Conflicted), vetoCount},
}
fmt.Print(tf.FormatRows(rows))
for _, v := range status.Vetoes {
fmt.Printf("  Veto: %s — %s\n", v.SummaryMessage, v.DetailedMessage)
}
return nil
},
}
}

func newPRDiffCmd(flags *GlobalFlags) *cobra.Command {
var contextLines int
var srcPath string

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
id, err := parsePRID(remaining)
if err != nil {
return err
}
contextLines = validation.ClampContextLines(contextLines)
diff, err := c.GetPullRequestDiff(context.Background(), project, repo, id, contextLines, srcPath)
if err != nil {
return err
}
if flags.JSON || flags.Format != "" {
return printFormatted(flags, diff)
}
fmt.Print(renderDiff(diff))
return nil
},
}
cmd.Flags().IntVar(&contextLines, "context", 10, "context lines")
cmd.Flags().StringVar(&srcPath, "src-path", "", "source path filter")
return cmd
}

func newPRDiffStatCmd(flags *GlobalFlags) *cobra.Command {
var limit, page int

cmd := &cobra.Command{
Use:     "diffstat [project] [repo] <id>",
Aliases: []string{"changes"},
Short:   "Show pull request diff statistics",
RunE: func(cmd *cobra.Command, args []string) error {
c, cfg, err := newClient(flags)
if err != nil {
return err
}
project, repo, remaining, err := resolveProjectRepo(cfg, args, 2)
if err != nil {
return err
}
id, err := parsePRID(remaining)
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
{Header: "PATH", Width: 50},
{Header: "TYPE", Width: 12},
{Header: "NODE_TYPE", Width: 12},
{Header: "CONTENT_ID", Width: 12},
}
tf := output.NewTableFormatter(cols, flags.NoColor)
var rows [][]string
for _, ch := range results.Values {
rows = append(rows, []string{ch.Path.ToString, ch.Type, ch.NodeType, ch.ContentID})
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
id, err := parsePRID(remaining)
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
{Header: "ID", Width: 12},
{Header: "DISPLAY_ID", Width: 10},
{Header: "MESSAGE", Width: 50},
{Header: "AUTHOR", Width: 20},
{Header: "TIMESTAMP", Width: 20},
}
tf := output.NewTableFormatter(cols, flags.NoColor)
var rows [][]string
for _, cm := range results.Values {
msg := strings.Split(cm.Message, "\n")[0]
rows = append(rows, []string{
cm.ID, cm.DisplayID, msg, cm.Author.Name, timeAgo(cm.AuthorTimestamp),
})
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
id, err := parsePRID(remaining)
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
cols := []output.Column{
{Header: "ID", Width: 8},
{Header: "ACTION", Width: 15},
{Header: "USER", Width: 20},
{Header: "CREATED", Width: 20},
}
tf := output.NewTableFormatter(cols, flags.NoColor)
var rows [][]string
for _, a := range results.Values {
rows = append(rows, []string{
strconv.Itoa(a.ID), a.Action, a.User.DisplayName, timeAgo(a.CreatedDate),
})
}
fmt.Print(tf.FormatRows(rows))
return nil
},
}
cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
cmd.Flags().IntVar(&page, "page", 1, "page number")
return cmd
}

func newPRParticipantsCmd(flags *GlobalFlags) *cobra.Command {
var limit, page int

cmd := &cobra.Command{
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
id, err := parsePRID(remaining)
if err != nil {
return err
}
start := (page - 1) * limit
results, err := c.ListPullRequestParticipants(context.Background(), project, repo, id, start, limit)
if err != nil {
return err
}
if flags.JSON || flags.Format != "" {
return printFormatted(flags, results.Values)
}
cols := []output.Column{
{Header: "USER", Width: 25},
{Header: "ROLE", Width: 12},
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
cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
cmd.Flags().IntVar(&page, "page", 1, "page number")
return cmd
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
id, err := parsePRID(remaining)
if err != nil {
return err
}
if err := c.WatchPullRequest(context.Background(), project, repo, id); err != nil {
return err
}
fmt.Fprintf(cmd.ErrOrStderr(), "Now watching PR #%d\n", id)
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
id, err := parsePRID(remaining)
if err != nil {
return err
}
if err := c.UnwatchPullRequest(context.Background(), project, repo, id); err != nil {
return err
}
fmt.Fprintf(cmd.ErrOrStderr(), "No longer watching PR #%d\n", id)
return nil
},
}
}

func newPRPublishCmd(flags *GlobalFlags) *cobra.Command {
return &cobra.Command{
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
id, err := parsePRID(remaining)
if err != nil {
return err
}
ctx := context.Background()
current, err := c.GetPullRequest(ctx, project, repo, id)
if err != nil {
return err
}
pr, err := c.PublishDraft(ctx, project, repo, id, current.Version)
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
}

func newPRConvertToDraftCmd(flags *GlobalFlags) *cobra.Command {
return &cobra.Command{
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
id, err := parsePRID(remaining)
if err != nil {
return err
}
ctx := context.Background()
current, err := c.GetPullRequest(ctx, project, repo, id)
if err != nil {
return err
}
pr, err := c.ConvertToDraft(ctx, project, repo, id, current.Version)
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
return fmt.Errorf("--target is required")
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
cmd.Flags().StringVar(&target, "target", "", "target branch (required)")
cmd.Flags().StringVar(&description, "description", "", "PR description")
cmd.Flags().StringSliceVar(&reviewers, "reviewer", nil, "reviewer usernames (repeatable)")
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
id, err := parsePRID(remaining)
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
id, err := parsePRID(remaining)
if err != nil {
return err
}
if !ConfirmDangerous("pull request", strconv.Itoa(id), confirm) {
return fmt.Errorf("deletion cancelled")
}
ctx := context.Background()
current, err := c.GetPullRequest(ctx, project, repo, id)
if err != nil {
return err
}
if err := c.DeletePullRequest(ctx, project, repo, id, current.Version); err != nil {
return err
}
fmt.Fprintf(cmd.ErrOrStderr(), "✓ Pull request #%d deleted\n", id)
return nil
},
}
cmd.Flags().BoolVar(&confirm, "confirm", false, "skip confirmation prompt")
return cmd
}
