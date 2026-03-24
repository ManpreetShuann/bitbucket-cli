package cmd

import (
"context"
"fmt"
"strings"

"github.com/manu/bb/internal/client"
"github.com/manu/bb/internal/output"
"github.com/manu/bb/internal/validation"
"github.com/spf13/cobra"
)

func NewUserCmd(flags *GlobalFlags) *cobra.Command {
cmd := &cobra.Command{
Use:   "user",
Short: "User operations",
}
cmd.AddCommand(newUserFindCmd(flags))
return cmd
}

func newUserFindCmd(flags *GlobalFlags) *cobra.Command {
var limit, page int
var allPages bool

cmd := &cobra.Command{
Use:   "find <query>",
Short: "Find users",
Args:  cobra.MinimumNArgs(1),
RunE: func(cmd *cobra.Command, args []string) error {
c, _, err := newClient(flags)
if err != nil {
return err
}
query := strings.Join(args, " ")
limit = validation.ClampLimit(limit)
ctx := context.Background()

var users []client.User
if allPages {
start := 0
for {
resp, err := c.FindUser(ctx, query, start, limit)
if err != nil {
return err
}
users = append(users, resp.Values...)
if resp.IsLastPage {
break
}
start = resp.NextPageStart
}
} else {
start := (page - 1) * limit
resp, err := c.FindUser(ctx, query, start, limit)
if err != nil {
return err
}
users = resp.Values
}

if flags.JSON || flags.Format != "" {
return printFormatted(flags, users)
}
cols := []output.Column{
{Header: "NAME", Width: 20},
{Header: "DISPLAY_NAME", Width: 25},
{Header: "EMAIL", Width: 30},
{Header: "SLUG", Width: 15},
{Header: "ACTIVE", Width: 8},
}
tf := output.NewTableFormatter(cols, flags.NoColor)
var rows [][]string
for _, u := range users {
active := "✗"
if u.Active {
active = "✓"
}
rows = append(rows, []string{u.Name, u.DisplayName, u.EmailAddress, u.Slug, active})
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
