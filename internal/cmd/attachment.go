package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func NewAttachmentCmd(flags *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attachment",
		Short: "Manage attachments",
	}
	cmd.AddCommand(newAttachmentGetCmd(flags))
	cmd.AddCommand(newAttachmentMetaCmd(flags))
	cmd.AddCommand(newAttachmentSaveMetaCmd(flags))
	cmd.AddCommand(newAttachmentDeleteCmd(flags))
	cmd.AddCommand(newAttachmentDeleteMetaCmd(flags))
	return cmd
}

func newAttachmentGetCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get [project] [repo] <attachment-id>",
		Short: "Get attachment content",
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
				return fmt.Errorf("attachment ID is required")
			}
			content, err := c.GetAttachment(context.Background(), project, repo, remaining[0])
			if err != nil {
				return err
			}
			fmt.Print(content)
			return nil
		},
	}
}

func newAttachmentMetaCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "meta [project] [repo] <attachment-id>",
		Short: "Get attachment metadata",
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
				return fmt.Errorf("attachment ID is required")
			}
			meta, err := c.GetAttachmentMetadata(context.Background(), project, repo, remaining[0])
			if err != nil {
				return err
			}
			return printFormatted(flags, meta)
		},
	}
}

func newAttachmentSaveMetaCmd(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "save-meta [project] [repo] <attachment-id>",
		Short: "Save attachment metadata",
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
				return fmt.Errorf("attachment ID is required")
			}
			metadata := map[string]any{}
			result, err := c.SaveAttachmentMetadata(context.Background(), project, repo, remaining[0], metadata)
			if err != nil {
				return err
			}
			return printFormatted(flags, result)
		},
	}
}

func newAttachmentDeleteCmd(flags *GlobalFlags) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete [project] [repo] <attachment-id>",
		Short: "Delete an attachment [dangerous]",
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
				return fmt.Errorf("attachment ID is required")
			}
			attachmentID := remaining[0]
			if !ConfirmDangerous("attachment", attachmentID, confirm) {
				return fmt.Errorf("deletion cancelled")
			}
			if err := c.DeleteAttachment(context.Background(), project, repo, attachmentID); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Attachment '%s' deleted\n", attachmentID)
			return nil
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "skip confirmation prompt")
	return cmd
}

func newAttachmentDeleteMetaCmd(flags *GlobalFlags) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete-meta [project] [repo] <attachment-id>",
		Short: "Delete attachment metadata [dangerous]",
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
				return fmt.Errorf("attachment ID is required")
			}
			attachmentID := remaining[0]
			if !ConfirmDangerous("attachment metadata", attachmentID, confirm) {
				return fmt.Errorf("deletion cancelled")
			}
			if err := c.DeleteAttachmentMetadata(context.Background(), project, repo, attachmentID); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Attachment metadata '%s' deleted\n", attachmentID)
			return nil
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "skip confirmation prompt")
	return cmd
}
