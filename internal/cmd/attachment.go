package cmd

import (
"context"
"encoding/json"
"fmt"
"os"

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
var outputPath string

cmd := &cobra.Command{
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
if outputPath != "" {
if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
return fmt.Errorf("failed to write file: %w", err)
}
fmt.Fprintf(cmd.ErrOrStderr(), "✓ Saved to %s\n", outputPath)
return nil
}
fmt.Print(content)
return nil
},
}
cmd.Flags().StringVar(&outputPath, "output", "", "file path to save content")
return cmd
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
if flags.JSON || flags.Format != "" {
return printFormatted(flags, meta)
}
for k, v := range meta {
fmt.Printf("%s: %v\n", k, v)
}
return nil
},
}
}

func newAttachmentSaveMetaCmd(flags *GlobalFlags) *cobra.Command {
var metadataStr, filePath string

cmd := &cobra.Command{
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

var metadata map[string]any
if filePath != "" {
data, err := os.ReadFile(filePath)
if err != nil {
return fmt.Errorf("failed to read file: %w", err)
}
if err := json.Unmarshal(data, &metadata); err != nil {
return fmt.Errorf("invalid JSON in file: %w", err)
}
} else if metadataStr != "" {
if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
return fmt.Errorf("invalid JSON: %w", err)
}
} else {
return fmt.Errorf("either --metadata or --file is required")
}

result, err := c.SaveAttachmentMetadata(context.Background(), project, repo, remaining[0], metadata)
if err != nil {
return err
}
return printFormatted(flags, result)
},
}
cmd.Flags().StringVar(&metadataStr, "metadata", "", "metadata as JSON string")
cmd.Flags().StringVar(&filePath, "file", "", "path to JSON file with metadata")
return cmd
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
