package cmd

import (
"context"
"fmt"
"strconv"

"github.com/ManpreetShuann/bitbucket-cli/internal/client"
"github.com/ManpreetShuann/bitbucket-cli/internal/output"
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
prID, err := parsePRID(remaining)
if err != nil {
return err
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
{Header: "TEXT", Width: 50},
{Header: "AUTHOR", Width: 15},
{Header: "STATE", Width: 10},
{Header: "SEVERITY", Width: 10},
}
tf := output.NewTableFormatter(cols, flags.NoColor)
var rows [][]string
for _, cm := range results.Values {
text := cm.Text
if len(text) > 50 {
text = text[:47] + "..."
}
rows = append(rows, []string{
strconv.Itoa(cm.ID), text, cm.Author.DisplayName, cm.State, cm.Severity,
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
prID, err := parsePRID(remaining)
if err != nil {
return err
}
if len(remaining) < 2 {
return fmt.Errorf("comment ID is required")
}
commentID, err := strconv.Atoi(remaining[1])
if err != nil {
return fmt.Errorf("invalid comment ID: %s", remaining[1])
}
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
var text, file, lineType, fileType string
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
prID, err := parsePRID(remaining)
if err != nil {
return err
}
if text == "" {
return fmt.Errorf("--text is required")
}
input := client.AddCommentInput{
Text:     text,
FilePath: file,
LineType: lineType,
FileType: fileType,
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
return cmd
}

func newPRCommentUpdateCmd(flags *GlobalFlags) *cobra.Command {
var text string

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
prID, err := parsePRID(remaining)
if err != nil {
return err
}
if len(remaining) < 2 {
return fmt.Errorf("comment ID is required")
}
commentID, err := strconv.Atoi(remaining[1])
if err != nil {
return fmt.Errorf("invalid comment ID: %s", remaining[1])
}
if text == "" {
return fmt.Errorf("--text is required")
}
ctx := context.Background()
current, err := c.GetPRComment(ctx, project, repo, prID, commentID)
if err != nil {
return err
}
comment, err := c.UpdatePRComment(ctx, project, repo, prID, commentID, current.Version, text)
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
return cmd
}

func newPRCommentResolveCmd(flags *GlobalFlags) *cobra.Command {
return &cobra.Command{
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
prID, err := parsePRID(remaining)
if err != nil {
return err
}
if len(remaining) < 2 {
return fmt.Errorf("comment ID is required")
}
commentID, err := strconv.Atoi(remaining[1])
if err != nil {
return fmt.Errorf("invalid comment ID: %s", remaining[1])
}
ctx := context.Background()
current, err := c.GetPRComment(ctx, project, repo, prID, commentID)
if err != nil {
return err
}
_, err = c.ResolvePRComment(ctx, project, repo, prID, commentID, current.Version)
if err != nil {
return err
}
fmt.Fprintf(cmd.ErrOrStderr(), "✓ Comment #%d resolved\n", commentID)
return nil
},
}
}

func newPRCommentReopenCmd(flags *GlobalFlags) *cobra.Command {
return &cobra.Command{
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
prID, err := parsePRID(remaining)
if err != nil {
return err
}
if len(remaining) < 2 {
return fmt.Errorf("comment ID is required")
}
commentID, err := strconv.Atoi(remaining[1])
if err != nil {
return fmt.Errorf("invalid comment ID: %s", remaining[1])
}
ctx := context.Background()
current, err := c.GetPRComment(ctx, project, repo, prID, commentID)
if err != nil {
return err
}
_, err = c.ReopenPRComment(ctx, project, repo, prID, commentID, current.Version)
if err != nil {
return err
}
fmt.Fprintf(cmd.ErrOrStderr(), "✓ Comment #%d reopened\n", commentID)
return nil
},
}
}

func newPRCommentDeleteCmd(flags *GlobalFlags) *cobra.Command {
var confirm bool

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
prID, err := parsePRID(remaining)
if err != nil {
return err
}
if len(remaining) < 2 {
return fmt.Errorf("comment ID is required")
}
commentID, err := strconv.Atoi(remaining[1])
if err != nil {
return fmt.Errorf("invalid comment ID: %s", remaining[1])
}
if !ConfirmDangerous("comment", strconv.Itoa(commentID), confirm) {
return fmt.Errorf("deletion cancelled")
}
ctx := context.Background()
current, err := c.GetPRComment(ctx, project, repo, prID, commentID)
if err != nil {
return err
}
if err := c.DeletePRComment(ctx, project, repo, prID, commentID, current.Version); err != nil {
return err
}
fmt.Fprintf(cmd.ErrOrStderr(), "✓ Comment #%d deleted\n", commentID)
return nil
},
}
cmd.Flags().BoolVar(&confirm, "confirm", false, "skip confirmation prompt")
return cmd
}
