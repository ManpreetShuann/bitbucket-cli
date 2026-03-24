package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manu/bb/internal/output"
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

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List dashboard pull requests",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, _, err := newClient(flags)
			if err != nil {
				return err
			}
			start := (page - 1) * limit
			results, err := c.ListDashboardPRs(context.Background(), state, role, order, start, limit)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, results.Values)
			}
			cols := []output.Column{
				{Header: "ID", Width: 6},
				{Header: "REPO", Width: 20},
				{Header: "TITLE", Width: 40},
				{Header: "AUTHOR", Width: 15},
				{Header: "STATE", Width: 10},
				{Header: "UPDATED", Width: 10},
			}
			tf := output.NewTableFormatter(cols, flags.NoColor)
			var rows [][]string
			for _, pr := range results.Values {
				repoSlug := ""
				if pr.ToRef.Repository.Slug != "" {
					repoSlug = pr.ToRef.Repository.Project.Key + "/" + pr.ToRef.Repository.Slug
				}
				rows = append(rows, []string{
					strconv.Itoa(pr.ID),
					repoSlug,
					pr.Title,
					pr.Author.User.DisplayName,
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
	cmd.Flags().StringVar(&state, "state", "", "filter by state")
	cmd.Flags().StringVar(&role, "role", "", "filter by role (AUTHOR, REVIEWER, PARTICIPANT)")
	cmd.Flags().StringVar(&order, "order", "", "OLDEST or NEWEST")
	return cmd
}

func newDashboardInboxCmd(flags *GlobalFlags) *cobra.Command {
	var limit, page int

	cmd := &cobra.Command{
		Use:   "inbox",
		Short: "List inbox pull requests",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, _, err := newClient(flags)
			if err != nil {
				return err
			}
			start := (page - 1) * limit
			results, err := c.ListInboxPRs(context.Background(), start, limit)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, results.Values)
			}
			cols := []output.Column{
				{Header: "ID", Width: 6},
				{Header: "REPO", Width: 20},
				{Header: "TITLE", Width: 40},
				{Header: "AUTHOR", Width: 15},
				{Header: "STATE", Width: 10},
			}
			tf := output.NewTableFormatter(cols, flags.NoColor)
			var rows [][]string
			for _, pr := range results.Values {
				repoSlug := ""
				if pr.ToRef.Repository.Slug != "" {
					repoSlug = pr.ToRef.Repository.Project.Key + "/" + pr.ToRef.Repository.Slug
				}
				rows = append(rows, []string{
					strconv.Itoa(pr.ID),
					repoSlug,
					pr.Title,
					pr.Author.User.DisplayName,
					pr.State,
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
