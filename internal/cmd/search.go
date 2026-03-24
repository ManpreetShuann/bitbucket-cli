package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/manu/bb/internal/output"
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
			start := (page - 1) * limit
			results, err := c.SearchCode(context.Background(), query, project, repo, start, limit)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, results)
			}
			if len(results.Results) == 0 {
				fmt.Println("No results found.")
				return nil
			}
			cols := []output.Column{
				{Header: "REPO", Width: 25},
				{Header: "FILE", Width: 40},
				{Header: "HITS", Width: 5},
			}
			tf := output.NewTableFormatter(cols, flags.NoColor)
			var rows [][]string
			for _, r := range results.Results {
				repoName := r.File.Repository.Slug
				rows = append(rows, []string{repoName, r.File.Path, fmt.Sprintf("%d", r.HitCount)})
			}
			fmt.Print(tf.FormatRows(rows))
			return nil
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "filter by project key")
	cmd.Flags().StringVar(&repo, "repo", "", "filter by repo slug")
	cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	return cmd
}
