package cmd

import (
"context"
"fmt"
"strings"
"time"

"github.com/ManpreetShuann/bitbucket-cli/internal/client"
"github.com/ManpreetShuann/bitbucket-cli/internal/output"
"github.com/ManpreetShuann/bitbucket-cli/internal/validation"
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
var allPages bool

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
limit = validation.ClampLimit(limit)
ctx := context.Background()

var commits []client.Commit
if allPages {
start := 0
for {
resp, err := c.ListCommits(ctx, project, repo, until, since, path, start, limit)
if err != nil {
return err
}
commits = append(commits, resp.Values...)
if resp.IsLastPage {
break
}
start = resp.NextPageStart
}
} else {
start := (page - 1) * limit
resp, err := c.ListCommits(ctx, project, repo, until, since, path, start, limit)
if err != nil {
return err
}
commits = resp.Values
}

if flags.JSON || flags.Format != "" {
return printFormatted(flags, commits)
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
for _, cm := range commits {
msg := strings.Split(cm.Message, "\n")[0]
ts := time.UnixMilli(cm.AuthorTimestamp).Format("2006-01-02 15:04")
id := cm.ID
if len(id) > 12 {
id = id[:12]
}
rows = append(rows, []string{id, cm.DisplayID, msg, cm.Author.Name, ts})
}
fmt.Print(tf.FormatRows(rows))
return nil
},
}
cmd.Flags().StringVar(&until, "until", "", "commit hash or ref upper bound")
cmd.Flags().StringVar(&since, "since", "", "commit hash or ref lower bound")
cmd.Flags().StringVar(&path, "path", "", "filter by file path")
cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
cmd.Flags().IntVar(&page, "page", 1, "page number")
cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
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
commitID := remaining[0]
if err := validation.ValidateCommitID(commitID); err != nil {
return err
}
commit, err := c.GetCommit(context.Background(), project, repo, commitID)
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
commitID := remaining[0]
if err := validation.ValidateCommitID(commitID); err != nil {
return err
}
contextLines = validation.ClampContextLines(contextLines)
diff, err := c.GetCommitDiff(context.Background(), project, repo, commitID, contextLines, srcPath)
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
cmd.Flags().StringVar(&srcPath, "src-path", "", "source file path")
return cmd
}

func newCommitChangesCmd(flags *GlobalFlags) *cobra.Command {
var limit, page int
var allPages bool

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
commitID := remaining[0]
if err := validation.ValidateCommitID(commitID); err != nil {
return err
}
limit = validation.ClampLimit(limit)
ctx := context.Background()

var changes []client.Change
if allPages {
start := 0
for {
resp, err := c.GetCommitChanges(ctx, project, repo, commitID, start, limit)
if err != nil {
return err
}
changes = append(changes, resp.Values...)
if resp.IsLastPage {
break
}
start = resp.NextPageStart
}
} else {
start := (page - 1) * limit
resp, err := c.GetCommitChanges(ctx, project, repo, commitID, start, limit)
if err != nil {
return err
}
changes = resp.Values
}

if flags.JSON || flags.Format != "" {
return printFormatted(flags, changes)
}
cols := []output.Column{
{Header: "PATH", Width: 50},
{Header: "TYPE", Width: 12},
{Header: "NODE_TYPE", Width: 12},
{Header: "CONTENT_ID", Width: 12},
}
tf := output.NewTableFormatter(cols, flags.NoColor)
var rows [][]string
for _, ch := range changes {
contentID := ch.ContentID
if len(contentID) > 12 {
contentID = contentID[:12]
}
rows = append(rows, []string{ch.Path.ToString, ch.Type, ch.NodeType, contentID})
}
fmt.Print(tf.FormatRows(rows))
return nil
},
}
cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
cmd.Flags().IntVar(&page, "page", 1, "page number")
cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
return cmd
}

func renderDiff(d *client.Diff) string {
var sb strings.Builder
for _, entry := range d.Diffs {
src := "(new file)"
if entry.Source != nil {
src = entry.Source.ToString
}
dst := "(deleted)"
if entry.Destination != nil {
dst = entry.Destination.ToString
}
sb.WriteString(fmt.Sprintf("--- a/%s\n+++ b/%s\n", src, dst))
for _, hunk := range entry.Hunks {
sb.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", hunk.SourceLine, hunk.SourceSpan, hunk.DestinationLine, hunk.DestinationSpan))
for _, seg := range hunk.Segments {
prefix := " "
if seg.Type == "ADDED" {
prefix = "+"
} else if seg.Type == "REMOVED" {
prefix = "-"
}
for _, line := range seg.Lines {
sb.WriteString(prefix + line.Line + "\n")
}
}
}
}
return sb.String()
}
