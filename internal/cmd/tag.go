package cmd

import (
"context"
"fmt"

"github.com/ManpreetShuann/bitbucket-cli/internal/client"
"github.com/ManpreetShuann/bitbucket-cli/internal/output"
"github.com/ManpreetShuann/bitbucket-cli/internal/validation"
"github.com/spf13/cobra"
)

func NewTagCmd(flags *GlobalFlags) *cobra.Command {
cmd := &cobra.Command{
Use:   "tag",
Short: "Manage tags",
}
cmd.AddCommand(newTagListCmd(flags))
cmd.AddCommand(newTagDeleteCmd(flags))
return cmd
}

func newTagListCmd(flags *GlobalFlags) *cobra.Command {
var limit, page int
var filter string
var allPages bool

cmd := &cobra.Command{
Use:   "list [project] [repo]",
Short: "List tags",
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

var tags []client.Tag
if allPages {
start := 0
for {
resp, err := c.ListTags(ctx, project, repo, filter, start, limit)
if err != nil {
return err
}
tags = append(tags, resp.Values...)
if resp.IsLastPage {
break
}
start = resp.NextPageStart
}
} else {
start := (page - 1) * limit
resp, err := c.ListTags(ctx, project, repo, filter, start, limit)
if err != nil {
return err
}
tags = resp.Values
}

if flags.JSON || flags.Format != "" {
return printFormatted(flags, tags)
}
cols := []output.Column{
{Header: "ID", Width: 40},
{Header: "DISPLAY_ID", Width: 25},
{Header: "LATEST_COMMIT", Width: 12},
{Header: "HASH", Width: 12},
}
tf := output.NewTableFormatter(cols, flags.NoColor)
var rows [][]string
for _, t := range tags {
commit := t.LatestCommit
if len(commit) > 12 {
commit = commit[:12]
}
hash := t.Hash
if len(hash) > 12 {
hash = hash[:12]
}
rows = append(rows, []string{t.ID, t.DisplayID, commit, hash})
}
fmt.Print(tf.FormatRows(rows))
return nil
},
}
cmd.Flags().StringVar(&filter, "filter", "", "filter tags by name")
cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
cmd.Flags().IntVar(&page, "page", 1, "page number")
cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
return cmd
}

func newTagDeleteCmd(flags *GlobalFlags) *cobra.Command {
var confirm bool

cmd := &cobra.Command{
Use:   "delete [project] [repo] <name>",
Short: "Delete a tag [dangerous]",
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
return fmt.Errorf("tag name is required")
}
name := remaining[0]
if !ConfirmDangerous("tag", name, confirm) {
return fmt.Errorf("deletion cancelled")
}
if err := c.DeleteTag(context.Background(), project, repo, name); err != nil {
return err
}
fmt.Fprintf(cmd.ErrOrStderr(), "✓ Tag '%s' deleted\n", name)
return nil
},
}
cmd.Flags().BoolVar(&confirm, "confirm", false, "skip confirmation prompt")
return cmd
}
