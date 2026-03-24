package cmd

import (
"context"
"fmt"
"strings"

"github.com/ManpreetShuann/bitbucket-cli/internal/client"
"github.com/ManpreetShuann/bitbucket-cli/internal/output"
"github.com/ManpreetShuann/bitbucket-cli/internal/validation"
"github.com/spf13/cobra"
)

func NewSearchCmd(flags *GlobalFlags) *cobra.Command {
cmd := &cobra.Command{
Use:   "search",
Short: "Search code",
}
cmd.AddCommand(newSearchCodeCmd(flags))
return cmd
}

func newSearchCodeCmd(flags *GlobalFlags) *cobra.Command {
var project, repo string
var limit, page int
var allPages bool

cmd := &cobra.Command{
Use:   "code <query>",
Short: "Search code across repositories",
Args:  cobra.MinimumNArgs(1),
RunE: func(cmd *cobra.Command, args []string) error {
c, _, err := newClient(flags)
if err != nil {
return err
}
query := strings.Join(args, " ")
limit = validation.ClampLimit(limit)
ctx := context.Background()

var allResults []client.SearchResult
if allPages {
start := 0
for {
results, err := c.SearchCode(ctx, query, project, repo, start, limit)
if err != nil {
return err
}
allResults = append(allResults, results.Results...)
if len(results.Results) < limit {
break
}
start += limit
}
} else {
start := (page - 1) * limit
results, err := c.SearchCode(ctx, query, project, repo, start, limit)
if err != nil {
return err
}
allResults = results.Results
}

if flags.JSON || flags.Format != "" {
return printFormatted(flags, allResults)
}
if len(allResults) == 0 {
fmt.Println("No results found.")
return nil
}
cols := []output.Column{
{Header: "FILE", Width: 50},
{Header: "REPOSITORY", Width: 30},
{Header: "HIT_COUNT", Width: 10},
}
tf := output.NewTableFormatter(cols, flags.NoColor)
var rows [][]string
for _, r := range allResults {
repoName := r.File.Repository.Slug
rows = append(rows, []string{r.File.Path, repoName, fmt.Sprintf("%d", r.HitCount)})
}
fmt.Print(tf.FormatRows(rows))
return nil
},
}
cmd.Flags().StringVar(&project, "project", "", "filter by project key")
cmd.Flags().StringVar(&repo, "repo", "", "filter by repo slug")
cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
cmd.Flags().IntVar(&page, "page", 1, "page number")
cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
return cmd
}
