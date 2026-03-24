package cmd

import (
	"context"
	"fmt"

	"github.com/manu/bb/internal/output"
	"github.com/spf13/cobra"
)

func NewFileCmd(flags *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "Browse and view files",
	}
	cmd.AddCommand(newFileBrowseCmd(flags))
	cmd.AddCommand(newFileCatCmd(flags))
	cmd.AddCommand(newFileListCmd(flags))
	cmd.AddCommand(newFileFindCmd(flags))
	return cmd
}

func newFileBrowseCmd(flags *GlobalFlags) *cobra.Command {
	var path, at string
	var limit, page int

	cmd := &cobra.Command{
		Use:   "browse [project] [repo]",
		Short: "Browse directory contents",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, _, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			start := (page - 1) * limit
			results, err := c.BrowseFiles(context.Background(), project, repo, path, at, start, limit)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, results.Values)
			}
			cols := []output.Column{
				{Header: "TYPE", Width: 6},
				{Header: "SIZE", Width: 10},
				{Header: "NAME", Width: 50},
			}
			tf := output.NewTableFormatter(cols, flags.NoColor)
			var rows [][]string
			for _, f := range results.Values {
				name := f.Path.ToString
				if name == "" {
					name = f.Path.Name
				}
				sizeStr := fmt.Sprintf("%d", f.Size)
				if f.Type == "DIRECTORY" {
					sizeStr = "-"
				}
				rows = append(rows, []string{f.Type[:3], sizeStr, name})
			}
			fmt.Print(tf.FormatRows(rows))
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "directory path")
	cmd.Flags().StringVar(&at, "at", "", "branch/tag/commit ref")
	cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	return cmd
}

func newFileCatCmd(flags *GlobalFlags) *cobra.Command {
	var at string

	cmd := &cobra.Command{
		Use:   "cat [project] [repo] <path>",
		Short: "Show file content",
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
				return fmt.Errorf("file path is required")
			}
			content, err := c.GetFileContent(context.Background(), project, repo, remaining[0], at)
			if err != nil {
				return err
			}
			fmt.Print(content)
			return nil
		},
	}
	cmd.Flags().StringVar(&at, "at", "", "branch/tag/commit ref")
	return cmd
}

func newFileListCmd(flags *GlobalFlags) *cobra.Command {
	var path, at string
	var limit, page int

	cmd := &cobra.Command{
		Use:   "list [project] [repo]",
		Short: "List files in a directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, _, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			start := (page - 1) * limit
			results, err := c.ListFiles(context.Background(), project, repo, path, at, start, limit)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, results.Values)
			}
			for _, f := range results.Values {
				fmt.Println(f)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "directory path")
	cmd.Flags().StringVar(&at, "at", "", "branch/tag/commit ref")
	cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	return cmd
}

func newFileFindCmd(flags *GlobalFlags) *cobra.Command {
	var limit, page int

	cmd := &cobra.Command{
		Use:   "find [project] [repo] <pattern>",
		Short: "Find files matching a pattern",
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
				return fmt.Errorf("search pattern is required")
			}
			start := (page - 1) * limit
			results, err := c.FindFile(context.Background(), project, repo, remaining[0], start, limit)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, results.Values)
			}
			for _, f := range results.Values {
				fmt.Println(f)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	return cmd
}
