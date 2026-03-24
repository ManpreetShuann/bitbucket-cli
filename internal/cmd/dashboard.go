package cmd

import (
"context"
"fmt"
"strconv"

"github.com/manu/bb/internal/client"
"github.com/manu/bb/internal/output"
"github.com/manu/bb/internal/validation"
"github.com/spf13/cobra"
)

func NewDashboardCmd(flags *GlobalFlags) *cobra.Command {
cmd := &cobra.Command{
Use:   "dashboard",
Short: "Dashboard and inbox",
}
cmd.AddCommand(newDashboardListCmd(flags))
cmd.AddCommand(newDashboardInboxCmd(flags))
return cmd
}

func newDashboardListCmd(flags *GlobalFlags) *cobra.Command {
var limit, page int
var state, role, order string
var allPages bool

cmd := &cobra.Command{
Use:   "list",
Short: "List dashboard pull requests",
RunE: func(cmd *cobra.Command, args []string) error {
c, _, err := newClient(flags)
if err != nil {
return err
}
limit = validation.ClampLimit(limit)
ctx := context.Background()

var prs []client.PullRequest
if allPages {
start := 0
for {
resp, err := c.ListDashboardPRs(ctx, state, role, order, start, limit)
if err != nil {
return err
}
prs = append(prs, resp.Values...)
if resp.IsLastPage {
break
}
start = resp.NextPageStart
}
} else {
start := (page - 1) * limit
resp, err := c.ListDashboardPRs(ctx, state, role, order, start, limit)
if err != nil {
return err
}
prs = resp.Values
}

if flags.JSON || flags.Format != "" {
return printFormatted(flags, prs)
}
cols := []output.Column{
{Header: "ID", Width: 8},
{Header: "TITLE", Width: 35},
{Header: "PROJECT/REPO", Width: 25},
{Header: "AUTHOR", Width: 15},
{Header: "STATE", Width: 10},
{Header: "REVIEWERS", Width: 20},
{Header: "UPDATED", Width: 15},
}
tf := output.NewTableFormatter(cols, flags.NoColor)
var rows [][]string
for _, pr := range prs {
repoSlug := ""
if pr.ToRef.Repository.Slug != "" {
repoSlug = pr.ToRef.Repository.Project.Key + "/" + pr.ToRef.Repository.Slug
}
rows = append(rows, []string{
strconv.Itoa(pr.ID),
pr.Title,
repoSlug,
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
cmd.Flags().StringVar(&state, "state", "OPEN", "filter by state")
cmd.Flags().StringVar(&role, "role", "", "filter by role (AUTHOR, REVIEWER, PARTICIPANT)")
cmd.Flags().StringVar(&order, "order", "", "OLDEST or NEWEST")
cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
cmd.Flags().IntVar(&page, "page", 1, "page number")
cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
return cmd
}

func newDashboardInboxCmd(flags *GlobalFlags) *cobra.Command {
var limit, page int
var allPages bool

cmd := &cobra.Command{
Use:   "inbox",
Short: "List inbox pull requests",
RunE: func(cmd *cobra.Command, args []string) error {
c, _, err := newClient(flags)
if err != nil {
return err
}
limit = validation.ClampLimit(limit)
ctx := context.Background()

var prs []client.PullRequest
if allPages {
start := 0
for {
resp, err := c.ListInboxPRs(ctx, start, limit)
if err != nil {
return err
}
prs = append(prs, resp.Values...)
if resp.IsLastPage {
break
}
start = resp.NextPageStart
}
} else {
start := (page - 1) * limit
resp, err := c.ListInboxPRs(ctx, start, limit)
if err != nil {
return err
}
prs = resp.Values
}

if flags.JSON || flags.Format != "" {
return printFormatted(flags, prs)
}
cols := []output.Column{
{Header: "ID", Width: 8},
{Header: "TITLE", Width: 40},
{Header: "PROJECT/REPO", Width: 25},
{Header: "STATE", Width: 10},
{Header: "UPDATED", Width: 15},
}
tf := output.NewTableFormatter(cols, flags.NoColor)
var rows [][]string
for _, pr := range prs {
repoSlug := ""
if pr.ToRef.Repository.Slug != "" {
repoSlug = pr.ToRef.Repository.Project.Key + "/" + pr.ToRef.Repository.Slug
}
rows = append(rows, []string{
strconv.Itoa(pr.ID),
pr.Title,
repoSlug,
pr.State,
timeAgo(pr.UpdatedDate),
})
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
