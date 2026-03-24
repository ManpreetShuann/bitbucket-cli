package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/manu/bb/internal/output"
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
			start := (page - 1) * limit
			results, err := c.FindUser(context.Background(), query, start, limit)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, results.Values)
			}
			cols := []output.Column{
				{Header: "USERNAME", Width: 20},
				{Header: "DISPLAY NAME", Width: 30},
				{Header: "EMAIL", Width: 30},
			}
			tf := output.NewTableFormatter(cols, flags.NoColor)
			var rows [][]string
			for _, u := range results.Values {
				rows = append(rows, []string{u.Name, u.DisplayName, u.EmailAddress})
			}
			fmt.Print(tf.FormatRows(rows))
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	return cmd
}
