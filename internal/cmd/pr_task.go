package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manu/bb/internal/output"
	"github.com/spf13/cobra"
)

func NewPRTaskCmd(flags *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage PR tasks",
	}
	cmd.AddCommand(newPRTaskListCmd(flags))
	cmd.AddCommand(newPRTaskCreateCmd(flags))
	cmd.AddCommand(newPRTaskGetCmd(flags))
	cmd.AddCommand(newPRTaskUpdateCmd(flags))
	cmd.AddCommand(newPRTaskDeleteCmd(flags))
	return cmd
}

func newPRTaskListCmd(flags *GlobalFlags) *cobra.Command {
	var limit, page int
	cmd := &cobra.Command{
		Use:   "list [project] [repo] <pr-id>",
		Short: "List PR tasks",
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
				return fmt.Errorf("PR ID is required")
			}
			prID, _ := strconv.Atoi(remaining[0])
			start := (page - 1) * limit
			results, err := c.ListPRTasks(context.Background(), project, repo, prID, start, limit)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, results.Values)
			}
			cols := []output.Column{
				{Header: "ID", Width: 8},
				{Header: "STATE", Width: 10},
				{Header: "TEXT", Width: 60},
			}
			tf := output.NewTableFormatter(cols, flags.NoColor)
			var rows [][]string
			for _, t := range results.Values {
				rows = append(rows, []string{strconv.Itoa(t.ID), t.State, t.Text})
			}
			fmt.Print(tf.FormatRows(rows))
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	return cmd
}

func newPRTaskCreateCmd(flags *GlobalFlags) *cobra.Command {
	var text string
	var commentID int

	cmd := &cobra.Command{
		Use:   "create [project] [repo] <pr-id>",
		Short: "Create a PR task",
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
				return fmt.Errorf("PR ID is required")
			}
			prID, _ := strconv.Atoi(remaining[0])
			if text == "" {
				return fmt.Errorf("--text is required")
			}
			var cid *int
			if cmd.Flags().Changed("comment-id") {
				cid = &commentID
			}
			task, err := c.CreatePRTask(context.Background(), project, repo, prID, text, cid)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, task)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Task #%d created\n", task.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&text, "text", "", "task text (required)")
	cmd.Flags().IntVar(&commentID, "comment-id", 0, "anchor to a comment")
	return cmd
}

func newPRTaskGetCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get [project] [repo] <pr-id> <task-id>",
		Short: "Get a PR task",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, remaining, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			if len(remaining) < 2 {
				return fmt.Errorf("PR ID and task ID are required")
			}
			prID, _ := strconv.Atoi(remaining[0])
			taskID, _ := strconv.Atoi(remaining[1])
			task, err := c.GetPRTask(context.Background(), project, repo, prID, taskID)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, task)
			}
			fmt.Printf("ID:    %d\n", task.ID)
			fmt.Printf("State: %s\n", task.State)
			fmt.Printf("Text:  %s\n", task.Text)
			return nil
		},
	}
}

func newPRTaskUpdateCmd(flags *GlobalFlags) *cobra.Command {
	var text, state string

	cmd := &cobra.Command{
		Use:   "update [project] [repo] <pr-id> <task-id>",
		Short: "Update a PR task",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, remaining, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			if len(remaining) < 2 {
				return fmt.Errorf("PR ID and task ID are required")
			}
			prID, _ := strconv.Atoi(remaining[0])
			taskID, _ := strconv.Atoi(remaining[1])
			task, err := c.UpdatePRTask(context.Background(), project, repo, prID, taskID, text, state)
			if err != nil {
				return err
			}
			if flags.JSON || flags.Format != "" {
				return printFormatted(flags, task)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Task #%d updated\n", task.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&text, "text", "", "new task text")
	cmd.Flags().StringVar(&state, "state", "", "task state (OPEN, RESOLVED)")
	return cmd
}

func newPRTaskDeleteCmd(flags *GlobalFlags) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete [project] [repo] <pr-id> <task-id>",
		Short: "Delete a PR task [dangerous]",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := newClient(flags)
			if err != nil {
				return err
			}
			project, repo, remaining, err := resolveProjectRepo(cfg, args, 2)
			if err != nil {
				return err
			}
			if len(remaining) < 2 {
				return fmt.Errorf("PR ID and task ID are required")
			}
			prID, _ := strconv.Atoi(remaining[0])
			taskID, _ := strconv.Atoi(remaining[1])
			name := fmt.Sprintf("task #%d", taskID)
			if !ConfirmDangerous("task", name, confirm) {
				return fmt.Errorf("deletion cancelled")
			}
			if err := c.DeletePRTask(context.Background(), project, repo, prID, taskID); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Task #%d deleted\n", taskID)
			return nil
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "skip confirmation prompt")
	return cmd
}
