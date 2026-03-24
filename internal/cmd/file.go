package cmd

import (
"context"
"fmt"

"github.com/manu/bb/internal/client"
"github.com/manu/bb/internal/output"
"github.com/manu/bb/internal/validation"
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
var allPages bool

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
limit = validation.ClampLimit(limit)
ctx := context.Background()

var entries []client.FileEntry
if allPages {
start := 0
for {
resp, err := c.BrowseFiles(ctx, project, repo, path, at, start, limit)
if err != nil {
return err
}
entries = append(entries, resp.Values...)
if resp.IsLastPage {
break
}
start = resp.NextPageStart
}
} else {
start := (page - 1) * limit
resp, err := c.BrowseFiles(ctx, project, repo, path, at, start, limit)
if err != nil {
return err
}
entries = resp.Values
}

if flags.JSON || flags.Format != "" {
return printFormatted(flags, entries)
}
cols := []output.Column{
{Header: "PATH", Width: 50},
{Header: "TYPE", Width: 12},
{Header: "SIZE", Width: 12},
{Header: "CONTENT_ID", Width: 12},
}
tf := output.NewTableFormatter(cols, flags.NoColor)
var rows [][]string
for _, f := range entries {
name := f.Path.ToString
if name == "" {
name = f.Path.Name
}
sizeStr := fmt.Sprintf("%d", f.Size)
if f.Type == "DIRECTORY" {
sizeStr = "-"
}
contentID := f.ContentID
if len(contentID) > 12 {
contentID = contentID[:12]
}
rows = append(rows, []string{name, f.Type, sizeStr, contentID})
}
fmt.Print(tf.FormatRows(rows))
return nil
},
}
cmd.Flags().StringVar(&path, "path", "", "directory path")
cmd.Flags().StringVar(&at, "at", "", "branch/tag/commit ref")
cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
cmd.Flags().IntVar(&page, "page", 1, "page number")
cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
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
filePath := remaining[0]
if err := validation.ValidatePath(filePath); err != nil {
return err
}
content, err := c.GetFileContent(context.Background(), project, repo, filePath, at)
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
var allPages bool

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
limit = validation.ClampLimit(limit)
ctx := context.Background()

var files []string
if allPages {
start := 0
for {
resp, err := c.ListFiles(ctx, project, repo, path, at, start, limit)
if err != nil {
return err
}
files = append(files, resp.Values...)
if resp.IsLastPage {
break
}
start = resp.NextPageStart
}
} else {
start := (page - 1) * limit
resp, err := c.ListFiles(ctx, project, repo, path, at, start, limit)
if err != nil {
return err
}
files = resp.Values
}

if flags.JSON || flags.Format != "" {
return printFormatted(flags, files)
}
cols := []output.Column{
{Header: "PATH", Width: 60},
}
tf := output.NewTableFormatter(cols, flags.NoColor)
var rows [][]string
for _, f := range files {
rows = append(rows, []string{f})
}
fmt.Print(tf.FormatRows(rows))
return nil
},
}
cmd.Flags().StringVar(&path, "path", "", "directory path")
cmd.Flags().StringVar(&at, "at", "", "branch/tag/commit ref")
cmd.Flags().IntVar(&limit, "limit", 25, "items per page")
cmd.Flags().IntVar(&page, "page", 1, "page number")
cmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")
return cmd
}

func newFileFindCmd(flags *GlobalFlags) *cobra.Command {
var limit, page int
var allPages bool

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
pattern := remaining[0]
limit = validation.ClampLimit(limit)
ctx := context.Background()

var files []string
if allPages {
start := 0
for {
resp, err := c.FindFile(ctx, project, repo, pattern, start, limit)
if err != nil {
return err
}
files = append(files, resp.Values...)
if resp.IsLastPage {
break
}
start = resp.NextPageStart
}
} else {
start := (page - 1) * limit
resp, err := c.FindFile(ctx, project, repo, pattern, start, limit)
if err != nil {
return err
}
files = resp.Values
}

if flags.JSON || flags.Format != "" {
return printFormatted(flags, files)
}
cols := []output.Column{
{Header: "PATH", Width: 60},
}
tf := output.NewTableFormatter(cols, flags.NoColor)
var rows [][]string
for _, f := range files {
rows = append(rows, []string{f})
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
