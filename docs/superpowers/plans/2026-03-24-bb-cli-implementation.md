# bb CLI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a complete Go CLI (`bb`) that wraps the Bitbucket Server REST API with 66 commands, profile-based auth, table/JSON/template output, and two-tier safety for dangerous operations.

**Architecture:** Flat client + command layer in `internal/`. HTTP client with generics-based pagination talks to Bitbucket REST API. Cobra command tree delegates to client methods and pipes results through output formatters. Viper handles config/profiles.

**Tech Stack:** Go 1.22+, Cobra, Viper, Lipgloss, stdlib `net/http`

**Spec:** `docs/superpowers/specs/2026-03-24-bb-cli-design.md`

**Reference implementation:** Python MCP server at `../bitbucket-server-mcp/` — every tool module maps to a CLI command group with exact API paths and parameters.

---

## File Structure

```
/home/manu/Code/bitbucket-cli/
├── cmd/bb/main.go                        # Entry point, version injection
├── internal/
│   ├── client/
│   │   ├── client.go                     # HTTP client struct, do(), retry, debug logging
│   │   ├── client_test.go                # httptest-based client tests
│   │   ├── errors.go                     # APIError type, exit codes
│   │   ├── pagination.go                 # PagedResponse[T], GetPaged, GetAll
│   │   ├── pagination_test.go
│   │   ├── types.go                      # All API data types (Project, Repo, PR, etc.)
│   │   ├── projects.go                   # ListProjects, GetProject
│   │   ├── repositories.go               # ListRepositories, GetRepository, CreateRepository
│   │   ├── branches.go                   # ListBranches, GetDefaultBranch, CreateBranch, ListTags
│   │   ├── pull_requests.go              # All PR API methods (CRUD, merge, approve, etc.)
│   │   ├── pull_request_comments.go      # PR comment API methods
│   │   ├── pull_request_tasks.go         # PR task API methods
│   │   ├── commits.go                    # ListCommits, GetCommit, GetCommitDiff, GetCommitChanges
│   │   ├── files.go                      # BrowseFiles, GetFileContent, ListFiles
│   │   ├── search.go                     # SearchCode, FindFile
│   │   ├── users.go                      # FindUser
│   │   ├── dashboard.go                  # ListDashboardPRs, ListInboxPRs
│   │   ├── attachments.go               # GetAttachment, GetAttachmentMetadata, SaveAttachmentMetadata
│   │   └── dangerous.go                  # Delete methods (branch, tag, PR, comment, task, attachment, project, repo)
│   ├── cmd/
│   │   ├── root.go                       # Root command, global flags, client init
│   │   ├── root_test.go
│   │   ├── auth.go                       # bb auth login/logout/status
│   │   ├── project.go                    # bb project list/get/delete
│   │   ├── repo.go                       # bb repo list/get/create/delete/use/clear
│   │   ├── branch.go                     # bb branch list/create/default/delete
│   │   ├── tag.go                        # bb tag list/delete
│   │   ├── pr.go                         # bb pr list/get/create/update/merge/decline/reopen/approve/...
│   │   ├── pr_comment.go                 # bb pr comment list/get/add/update/resolve/reopen/delete
│   │   ├── pr_task.go                    # bb pr task list/create/get/update/delete
│   │   ├── commit.go                     # bb commit list/get/diff/changes
│   │   ├── file.go                       # bb file browse/cat/list/find
│   │   ├── search.go                     # bb search code
│   │   ├── user.go                       # bb user find
│   │   ├── dashboard.go                  # bb dashboard list/inbox
│   │   ├── attachment.go                 # bb attachment get/meta/save-meta/delete/delete-meta
│   │   └── confirm.go                    # Shared confirmation prompts (dangerous/destructive)
│   ├── config/
│   │   ├── config.go                     # Profile loading, resolution order
│   │   ├── config_test.go
│   │   ├── auth.go                       # Credential read/write, file permissions
│   │   └── auth_test.go
│   ├── output/
│   │   ├── formatter.go                  # Formatter interface, NewFormatter factory
│   │   ├── formatter_test.go
│   │   ├── table.go                      # TableFormatter with Lipgloss
│   │   ├── json.go                       # JSONFormatter
│   │   └── template.go                   # TemplateFormatter (Go text/template)
│   └── validation/
│       ├── validation.go                 # All validators and clamp functions
│       └── validation_test.go
├── go.mod
├── go.sum
├── Makefile
└── .gitignore
```

---

## Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`, `cmd/bb/main.go`, `Makefile`, `.gitignore`

- [ ] **Step 1: Initialize Go module**

```bash
cd /home/manu/Code/bitbucket-cli
go mod init github.com/manu/bb
```

- [ ] **Step 2: Create main.go**

Create `cmd/bb/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/manu/bb/internal/cmd"
)

var version = "dev"

func main() {
	rootCmd := cmd.NewRootCmd(version)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 3: Create minimal root command**

Create `internal/cmd/root.go`:

```go
package cmd

import (
	"github.com/spf13/cobra"
)

type GlobalFlags struct {
	Profile string
	JSON    bool
	Format  string
	NoColor bool
	Debug   bool
}

func NewRootCmd(version string) *cobra.Command {
	flags := &GlobalFlags{}

	rootCmd := &cobra.Command{
		Use:     "bb",
		Short:   "Bitbucket Server CLI",
		Long:    "A command-line interface for Bitbucket Server",
		Version: version,
	}

	rootCmd.PersistentFlags().StringVar(&flags.Profile, "profile", "default", "config profile to use")
	rootCmd.PersistentFlags().BoolVar(&flags.JSON, "json", false, "output as JSON")
	rootCmd.PersistentFlags().StringVar(&flags.Format, "format", "", "Go template for custom output")
	rootCmd.PersistentFlags().BoolVar(&flags.NoColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&flags.Debug, "debug", false, "print HTTP request/response details to stderr")

	return rootCmd
}
```

- [ ] **Step 4: Create Makefile**

Create `Makefile`:

```makefile
BINARY    := bb
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -s -w -X main.version=$(VERSION)

.PHONY: build install test lint clean

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/bb

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/bb

test:
	go test ./... -race -cover

lint:
	golangci-lint run

clean:
	rm -rf bin/
```

- [ ] **Step 5: Create .gitignore**

```
bin/
*.exe
.DS_Store
```

- [ ] **Step 6: Install dependencies**

```bash
cd /home/manu/Code/bitbucket-cli
go get github.com/spf13/cobra
go get github.com/spf13/viper
go get github.com/charmbracelet/lipgloss
go get golang.org/x/term
```

- [ ] **Step 7: Verify build**

```bash
cd /home/manu/Code/bitbucket-cli && make build
bin/bb --version
# Expected: bb version dev
```

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "feat: scaffold bb CLI project with Go module, main, root command, and Makefile"
```

---

## Task 2: Validation Package

**Files:**
- Create: `internal/validation/validation.go`, `internal/validation/validation_test.go`

- [ ] **Step 1: Write validation tests**

Create `internal/validation/validation_test.go`:

```go
package validation

import (
	"strings"
	"testing"
)

func TestValidateProjectKey(t *testing.T) {
	valid := []string{"PROJ", "my_proj", "~jsmith", "A", strings.Repeat("a", 128)}
	for _, v := range valid {
		if err := ValidateProjectKey(v); err != nil {
			t.Errorf("ValidateProjectKey(%q) = %v, want nil", v, err)
		}
	}

	invalid := []string{"", "has spaces", "special@char", strings.Repeat("a", 129), "proj/key"}
	for _, v := range invalid {
		if err := ValidateProjectKey(v); err == nil {
			t.Errorf("ValidateProjectKey(%q) = nil, want error", v)
		}
	}
}

func TestValidateRepoSlug(t *testing.T) {
	valid := []string{"my-repo", "repo123", "my.repo", "A"}
	for _, v := range valid {
		if err := ValidateRepoSlug(v); err != nil {
			t.Errorf("ValidateRepoSlug(%q) = %v, want nil", v, err)
		}
	}

	invalid := []string{"", "-starts-with-dash", ".starts-with-dot", "has spaces"}
	for _, v := range invalid {
		if err := ValidateRepoSlug(v); err == nil {
			t.Errorf("ValidateRepoSlug(%q) = nil, want error", v)
		}
	}
}

func TestValidateCommitID(t *testing.T) {
	valid := []string{"abcd", "abc123def456", "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"}
	for _, v := range valid {
		if err := ValidateCommitID(v); err != nil {
			t.Errorf("ValidateCommitID(%q) = %v, want nil", v, err)
		}
	}

	invalid := []string{"", "abc", "xyz123", "not-hex-chars!"}
	for _, v := range invalid {
		if err := ValidateCommitID(v); err == nil {
			t.Errorf("ValidateCommitID(%q) = nil, want error", v)
		}
	}
}

func TestValidateBranchName(t *testing.T) {
	valid := []string{"main", "feature/my-branch", "release/v1.0", "a"}
	for _, v := range valid {
		if err := ValidateBranchName(v); err != nil {
			t.Errorf("ValidateBranchName(%q) = %v, want nil", v, err)
		}
	}

	invalid := []string{"", "-bad", "has//double", "trailing/", "../traversal", strings.Repeat("a", 257)}
	for _, v := range invalid {
		if err := ValidateBranchName(v); err == nil {
			t.Errorf("ValidateBranchName(%q) = nil, want error", v)
		}
	}
}

func TestValidatePath(t *testing.T) {
	valid := []string{"", "src/main.go", "path/to/file"}
	for _, v := range valid {
		if err := ValidatePath(v); err != nil {
			t.Errorf("ValidatePath(%q) = %v, want nil", v, err)
		}
	}

	invalid := []string{"/absolute", "path/../traversal", "has\x00null"}
	for _, v := range invalid {
		if err := ValidatePath(v); err == nil {
			t.Errorf("ValidatePath(%q) = nil, want error", v)
		}
	}
}

func TestValidateEnum(t *testing.T) {
	tests := []struct {
		fn    func(string) error
		valid []string
		bad   string
	}{
		{ValidatePRState, []string{"OPEN", "DECLINED", "MERGED", "ALL", "open"}, "INVALID"},
		{ValidatePRDirection, []string{"INCOMING", "OUTGOING", "incoming"}, "INVALID"},
		{ValidatePRRole, []string{"AUTHOR", "REVIEWER", "PARTICIPANT"}, "INVALID"},
		{ValidatePROrder, []string{"OLDEST", "NEWEST"}, "INVALID"},
		{ValidateTaskState, []string{"OPEN", "RESOLVED"}, "INVALID"},
	}
	for _, tt := range tests {
		for _, v := range tt.valid {
			if err := tt.fn(v); err != nil {
				t.Errorf("validate(%q) = %v, want nil", v, err)
			}
		}
		if err := tt.fn(tt.bad); err == nil {
			t.Errorf("validate(%q) = nil, want error", tt.bad)
		}
	}
}

func TestClampLimit(t *testing.T) {
	tests := []struct{ in, want int }{
		{0, 1}, {1, 1}, {25, 25}, {1000, 1000}, {1001, 1000}, {-5, 1},
	}
	for _, tt := range tests {
		if got := ClampLimit(tt.in); got != tt.want {
			t.Errorf("ClampLimit(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestClampContextLines(t *testing.T) {
	tests := []struct{ in, want int }{
		{-1, 0}, {0, 0}, {10, 10}, {100, 100}, {101, 100},
	}
	for _, tt := range tests {
		if got := ClampContextLines(tt.in); got != tt.want {
			t.Errorf("ClampContextLines(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /home/manu/Code/bitbucket-cli && go test ./internal/validation/ -v
# Expected: FAIL — package doesn't exist yet
```

- [ ] **Step 3: Implement validation.go**

Create `internal/validation/validation.go`:

```go
package validation

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	projectKeyRE = regexp.MustCompile(`^~?[A-Za-z0-9_]{1,128}$`)
	repoSlugRE   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
	commitIDRE   = regexp.MustCompile(`^[0-9a-fA-F]{4,40}$`)
	branchNameRE = regexp.MustCompile(`^(?:.*//)`)
	tagNameRE    = branchNameRE // Same rules

	validPRStates      = map[string]bool{"OPEN": true, "DECLINED": true, "MERGED": true, "ALL": true}
	validPRDirections  = map[string]bool{"INCOMING": true, "OUTGOING": true}
	validPRRoles       = map[string]bool{"AUTHOR": true, "REVIEWER": true, "PARTICIPANT": true}
	validPROrders      = map[string]bool{"OLDEST": true, "NEWEST": true}
	validTaskStates    = map[string]bool{"OPEN": true, "RESOLVED": true}
)

const (
	MaxLimit        = 1000
	MaxContextLines = 100
	MaxBranchLen    = 256
)

func ValidateProjectKey(key string) error {
	if !projectKeyRE.MatchString(key) {
		return fmt.Errorf("invalid project key: %q (must be alphanumeric/underscores, 1-128 chars, optional ~ prefix)", key)
	}
	return nil
}

func ValidateRepoSlug(slug string) error {
	if !repoSlugRE.MatchString(slug) {
		return fmt.Errorf("invalid repo slug: %q (must start with alphanumeric, contain only [A-Za-z0-9._-])", slug)
	}
	return nil
}

func ValidateCommitID(id string) error {
	if !commitIDRE.MatchString(id) {
		return fmt.Errorf("invalid commit ID: %q (must be a hex SHA, 4-40 chars)", id)
	}
	return nil
}

func ValidateBranchName(name string) error {
	if name == "" {
		return fmt.Errorf("branch name must not be empty")
	}
	if len(name) > MaxBranchLen {
		return fmt.Errorf("branch name too long: %d chars (max %d)", len(name), MaxBranchLen)
	}
	if name[0] < '0' || (name[0] > '9' && name[0] < 'A') || (name[0] > 'Z' && name[0] < 'a') || name[0] > 'z' {
		return fmt.Errorf("branch name must start with alphanumeric character: %q", name)
	}
	if strings.Contains(name, "//") {
		return fmt.Errorf("branch name must not contain '//': %q", name)
	}
	if strings.HasSuffix(name, "/") {
		return fmt.Errorf("branch name must not end with '/': %q", name)
	}
	for _, seg := range strings.Split(name, "/") {
		if seg == ".." {
			return fmt.Errorf("branch name must not contain path traversal: %q", name)
		}
	}
	return nil
}

func ValidateTagName(name string) error {
	return ValidateBranchName(name) // Same rules
}

func ValidatePath(path string) error {
	if path == "" {
		return nil
	}
	if strings.ContainsRune(path, 0) {
		return fmt.Errorf("path must not contain null bytes")
	}
	if strings.HasPrefix(path, "/") {
		return fmt.Errorf("path must not start with '/'")
	}
	for _, seg := range strings.Split(path, "/") {
		if seg == ".." {
			return fmt.Errorf("path traversal ('..') is not permitted")
		}
	}
	return nil
}

func ValidatePositiveInt(value int, name string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be a positive integer, got %d", name, value)
	}
	return nil
}

func validateEnum(value string, valid map[string]bool, name string) error {
	upper := strings.ToUpper(value)
	if !valid[upper] {
		keys := make([]string, 0, len(valid))
		for k := range valid {
			keys = append(keys, k)
		}
		return fmt.Errorf("invalid %s: %q (must be one of %v)", name, value, keys)
	}
	return nil
}

func ValidatePRState(state string) error     { return validateEnum(state, validPRStates, "PR state") }
func ValidatePRDirection(dir string) error    { return validateEnum(dir, validPRDirections, "PR direction") }
func ValidatePRRole(role string) error        { return validateEnum(role, validPRRoles, "PR role") }
func ValidatePROrder(order string) error      { return validateEnum(order, validPROrders, "PR order") }
func ValidateTaskState(state string) error    { return validateEnum(state, validTaskStates, "task state") }

func ClampLimit(limit int) int {
	if limit < 1 {
		return 1
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}

func ClampStart(start int) int {
	if start < 0 {
		return 0
	}
	return start
}

func ClampContextLines(lines int) int {
	if lines < 0 {
		return 0
	}
	if lines > MaxContextLines {
		return MaxContextLines
	}
	return lines
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /home/manu/Code/bitbucket-cli && go test ./internal/validation/ -v
# Expected: all PASS
```

- [ ] **Step 5: Commit**

```bash
git add internal/validation/
git commit -m "feat: add input validation package with all validators and clamp functions"
```

---

## Task 3: API Types

**Files:**
- Create: `internal/client/types.go`

- [ ] **Step 1: Create types.go**

Create `internal/client/types.go` with all Bitbucket API data types. Every field needs a JSON struct tag matching the Bitbucket REST API field names exactly.

```go
package client

// Project represents a Bitbucket Server project.
type Project struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Public      bool   `json:"public"`
	Type        string `json:"type"`
	Links       Links  `json:"links"`
}

// Repository represents a Bitbucket Server repository.
type Repository struct {
	Slug        string  `json:"slug"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	State       string  `json:"state"`
	ScmID       string  `json:"scmId"`
	Project     Project `json:"project"`
	Forkable    bool    `json:"forkable"`
	Links       Links   `json:"links"`
}

// PullRequest represents a Bitbucket Server pull request.
type PullRequest struct {
	ID          int           `json:"id"`
	Version     int           `json:"version"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	State       string        `json:"state"`
	Draft       bool          `json:"draft"`
	Author      Participant   `json:"author"`
	Reviewers   []Participant `json:"reviewers"`
	FromRef     Ref           `json:"fromRef"`
	ToRef       Ref           `json:"toRef"`
	CreatedDate int64         `json:"createdDate"`
	UpdatedDate int64         `json:"updatedDate"`
	Links       Links         `json:"links"`
}

// Participant represents a PR participant with role and approval status.
type Participant struct {
	User     User   `json:"user"`
	Role     string `json:"role"`
	Approved bool   `json:"approved"`
	Status   string `json:"status"`
}

// User represents a Bitbucket Server user.
type User struct {
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
	Slug         string `json:"slug"`
	Active       bool   `json:"active"`
	Links        Links  `json:"links"`
}

// Ref represents a git reference (branch/tag).
type Ref struct {
	ID           string     `json:"id"`
	DisplayID    string     `json:"displayId"`
	LatestCommit string     `json:"latestCommit"`
	Repository   Repository `json:"repository"`
}

// Branch represents a branch in a repository.
type Branch struct {
	ID              string `json:"id"`
	DisplayID       string `json:"displayId"`
	LatestCommit    string `json:"latestCommit"`
	IsDefault       bool   `json:"isDefault"`
}

// Tag represents a git tag.
type Tag struct {
	ID           string `json:"id"`
	DisplayID    string `json:"displayId"`
	LatestCommit string `json:"latestCommit"`
	Hash         string `json:"hash"`
}

// Commit represents a git commit.
type Commit struct {
	ID              string   `json:"id"`
	DisplayID       string   `json:"displayId"`
	Message         string   `json:"message"`
	Author          Person   `json:"author"`
	AuthorTimestamp int64    `json:"authorTimestamp"`
	Committer       Person   `json:"committer"`
	Parents         []Commit `json:"parents"`
}

// Person represents a git author/committer.
type Person struct {
	Name         string `json:"name"`
	EmailAddress string `json:"emailAddress"`
}

// Comment represents a PR comment.
type Comment struct {
	ID          int       `json:"id"`
	Version     int       `json:"version"`
	Text        string    `json:"text"`
	Author      User      `json:"author"`
	Severity    string    `json:"severity"`
	State       string    `json:"state"`
	Anchor      *Anchor   `json:"anchor,omitempty"`
	Comments    []Comment `json:"comments"`
	CreatedDate int64     `json:"createdDate"`
	UpdatedDate int64     `json:"updatedDate"`
}

// Anchor represents the file/line location of an inline comment.
type Anchor struct {
	Path     string `json:"path"`
	Line     int    `json:"line"`
	LineType string `json:"lineType"`
	FileType string `json:"fileType"`
}

// Task represents a PR task.
type Task struct {
	ID    int    `json:"id"`
	Text  string `json:"text"`
	State string `json:"state"`
}

// Activity represents a PR activity feed entry.
type Activity struct {
	ID          int    `json:"id"`
	Action      string `json:"action"`
	Comment     *Comment `json:"comment,omitempty"`
	CreatedDate int64  `json:"createdDate"`
	User        User   `json:"user"`
}

// Links holds HATEOAS links from the API.
type Links struct {
	Clone []Link `json:"clone"`
	Self  []Link `json:"self"`
}

// Link is a single HATEOAS link.
type Link struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

// MergeStatus represents the merge readiness check result.
type MergeStatus struct {
	CanMerge   bool   `json:"canMerge"`
	Conflicted bool   `json:"conflicted"`
	Vetoes     []Veto `json:"vetoes"`
}

// Veto represents a merge veto reason.
type Veto struct {
	SummaryMessage  string `json:"summaryMessage"`
	DetailedMessage string `json:"detailedMessage"`
}

// FileEntry represents a file or directory in a browse response.
type FileEntry struct {
	Path      FileEntryPath `json:"path,omitempty"`
	Type      string        `json:"type"`
	Size      int64         `json:"size"`
	ContentID string        `json:"contentId"`
}

// FileEntryPath contains the path components of a file entry.
type FileEntryPath struct {
	Components []string `json:"components"`
	Name       string   `json:"name"`
	ToString   string   `json:"toString"`
}

// Change represents a file changed in a commit.
type Change struct {
	ContentID  string         `json:"contentId"`
	FromHash   string         `json:"fromHash"`
	ToHash     string         `json:"toHash"`
	Path       FileEntryPath  `json:"path"`
	SrcPath    *FileEntryPath `json:"srcPath,omitempty"`
	Type       string         `json:"type"`
	NodeType   string         `json:"nodeType"`
}

// SearchResult represents a code search result.
type SearchResult struct {
	File       SearchFile      `json:"file,omitempty"`
	HitCount   int             `json:"hitCount"`
	PathMatches []SearchMatch  `json:"pathMatches,omitempty"`
	HitContexts []HitContext   `json:"hitContexts,omitempty"`
}

// SearchFile represents the file info in a search result.
type SearchFile struct {
	Path       string     `json:"path"`
	Repository Repository `json:"repository"`
}

// SearchMatch represents a matched path segment.
type SearchMatch struct {
	Text  string `json:"text"`
	Match bool   `json:"match"`
}

// HitContext represents surrounding lines of a code search hit.
type HitContext struct {
	Lines []HitLine `json:"lines"`
}

// HitLine represents a single line in a code search hit context.
type HitLine struct {
	Text string `json:"text"`
	Line int    `json:"line"`
}

// Diff represents a diff response.
type Diff struct {
	Diffs []DiffEntry `json:"diffs"`
}

// DiffEntry represents a single file diff.
type DiffEntry struct {
	Source      *DiffPath   `json:"source,omitempty"`
	Destination *DiffPath  `json:"destination,omitempty"`
	Hunks       []Hunk     `json:"hunks"`
	Truncated   bool       `json:"truncated"`
}

// DiffPath represents a path in a diff.
type DiffPath struct {
	ToString string `json:"toString"`
}

// Hunk represents a diff hunk.
type Hunk struct {
	SourceLine      int        `json:"sourceLine"`
	SourceSpan      int        `json:"sourceSpan"`
	DestinationLine int        `json:"destinationLine"`
	DestinationSpan int        `json:"destinationSpan"`
	Segments        []Segment  `json:"segments"`
}

// Segment represents a diff segment (context, added, removed).
type Segment struct {
	Type  string     `json:"type"`
	Lines []DiffLine `json:"lines"`
}

// DiffLine represents a single line in a diff.
type DiffLine struct {
	Source      int    `json:"source"`
	Destination int    `json:"destination"`
	Line        string `json:"line"`
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /home/manu/Code/bitbucket-cli && go build ./internal/client/
# Expected: Success (no output)
```

- [ ] **Step 3: Commit**

```bash
git add internal/client/types.go
git commit -m "feat: add all Bitbucket API data types"
```

---

## Task 4: HTTP Client Core

**Files:**
- Create: `internal/client/client.go`, `internal/client/errors.go`, `internal/client/client_test.go`

- [ ] **Step 1: Write client tests**

Create `internal/client/client_test.go`:

```go
package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/1.0/projects" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected auth: %s", r.Header.Get("Authorization"))
		}
		if r.Method != "GET" {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"key": "PROJ"})
	}))
	defer server.Close()

	c := New(server.URL, "test-token")
	var result map[string]any
	err := c.Get(context.Background(), "/projects", nil, &result)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if result["key"] != "PROJ" {
		t.Errorf("expected key=PROJ, got %v", result["key"])
	}
}

func TestClient_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("unexpected method: %s", r.Method)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "my-repo" {
			t.Errorf("unexpected body: %v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]any{"slug": "my-repo"})
	}))
	defer server.Close()

	c := New(server.URL, "test-token")
	body := map[string]any{"name": "my-repo"}
	var result map[string]any
	err := c.Post(context.Background(), "/projects/PROJ/repos", body, nil, &result)
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if result["slug"] != "my-repo" {
		t.Errorf("expected slug=my-repo, got %v", result["slug"])
	}
}

func TestClient_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]any{{"message": "Project not found"}},
		})
	}))
	defer server.Close()

	c := New(server.URL, "test-token")
	var result map[string]any
	err := c.Get(context.Background(), "/projects/NOPE", nil, &result)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
}

func TestClient_Retry429(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(429)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	c := New(server.URL, "test-token")
	var result map[string]any
	err := c.Get(context.Background(), "/test", nil, &result)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClient_204NoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer server.Close()

	c := New(server.URL, "test-token")
	var result map[string]any
	err := c.Post(context.Background(), "/test", nil, nil, &result)
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /home/manu/Code/bitbucket-cli && go test ./internal/client/ -v -run TestClient
# Expected: FAIL — functions don't exist
```

- [ ] **Step 3: Create errors.go**

Create `internal/client/errors.go`:

```go
package client

import (
	"fmt"
	"strings"
)

const (
	ExitSuccess    = 0
	ExitError      = 1
	ExitAuthError  = 2
	ExitNotFound   = 3
	ExitValidation = 4
)

type APIError struct {
	StatusCode int
	Errors     []APIErrorDetail
}

type APIErrorDetail struct {
	Context       string `json:"context"`
	Message       string `json:"message"`
	ExceptionName string `json:"exceptionName"`
}

func (e *APIError) Error() string {
	if len(e.Errors) > 0 {
		msgs := make([]string, len(e.Errors))
		for i, err := range e.Errors {
			msgs[i] = err.Message
		}
		return fmt.Sprintf("Bitbucket API error (%d): %s", e.StatusCode, strings.Join(msgs, "; "))
	}
	return fmt.Sprintf("Bitbucket API error (%d)", e.StatusCode)
}

func (e *APIError) ExitCode() int {
	switch {
	case e.StatusCode == 401 || e.StatusCode == 403:
		return ExitAuthError
	case e.StatusCode == 404:
		return ExitNotFound
	default:
		return ExitError
	}
}
```

- [ ] **Step 4: Create client.go**

Create `internal/client/client.go`:

```go
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	debug      bool
}

func New(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}

func (c *Client) SetTimeout(d time.Duration) {
	c.httpClient.Timeout = d
}

// Get performs a GET to /rest/api/1.0{path}.
func (c *Client) Get(ctx context.Context, path string, params url.Values, result any) error {
	return c.do(ctx, "GET", "/rest/api/1.0"+path, params, nil, result)
}

// Post performs a POST to /rest/api/1.0{path}.
func (c *Client) Post(ctx context.Context, path string, body any, params url.Values, result any) error {
	return c.do(ctx, "POST", "/rest/api/1.0"+path, params, body, result)
}

// Put performs a PUT to /rest/api/1.0{path}.
func (c *Client) Put(ctx context.Context, path string, body any, params url.Values, result any) error {
	return c.do(ctx, "PUT", "/rest/api/1.0"+path, params, body, result)
}

// Delete performs a DELETE to /rest/api/1.0{path}.
func (c *Client) Delete(ctx context.Context, path string, params url.Values, result any) error {
	return c.do(ctx, "DELETE", "/rest/api/1.0"+path, params, nil, result)
}

// DoAbsolute performs a request to an absolute path (not prefixed with /rest/api/1.0).
func (c *Client) DoAbsolute(ctx context.Context, method, path string, body any, params url.Values, result any) error {
	return c.do(ctx, method, path, params, body, result)
}

// GetRaw fetches raw file content as a string (not JSON).
func (c *Client) GetRaw(ctx context.Context, path string, params url.Values) (string, error) {
	fullURL := c.baseURL + "/rest/api/1.0" + path
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return "", err
	}
	c.setHeaders(req)

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	c.debugLog("GET", path, resp.StatusCode, time.Since(start))

	if resp.StatusCode >= 400 {
		return "", c.parseError(resp)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Search calls the code-search API at /rest/search/latest/search.
// Tries POST first, falls back to GET on 405.
func (c *Client) Search(ctx context.Context, body any, params url.Values, result any) error {
	err := c.do(ctx, "POST", "/rest/search/latest/search", nil, body, result)
	if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == 405 {
		return c.do(ctx, "GET", "/rest/search/latest/search", params, nil, result)
	}
	return err
}

func (c *Client) do(ctx context.Context, method, path string, params url.Values, body any, result any) error {
	var lastErr error
	delays := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

	for attempt := 0; attempt <= len(delays); attempt++ {
		if attempt > 0 {
			time.Sleep(delays[attempt-1])
		}

		err := c.doOnce(ctx, method, path, params, body, result)
		if err == nil {
			return nil
		}

		apiErr, ok := err.(*APIError)
		if !ok || (apiErr.StatusCode != 429 && apiErr.StatusCode != 503) {
			return err
		}
		lastErr = err
		if attempt >= len(delays) {
			break
		}
	}
	return lastErr
}

func (c *Client) doOnce(ctx context.Context, method, path string, params url.Values, body any, result any) error {
	fullURL := c.baseURL + path
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return err
	}
	c.setHeaders(req)

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	c.debugLog(method, path, resp.StatusCode, time.Since(start))

	if resp.StatusCode >= 400 {
		return c.parseError(resp)
	}

	if resp.StatusCode == 204 || result == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
}

func (c *Client) parseError(resp *http.Response) error {
	apiErr := &APIError{StatusCode: resp.StatusCode}

	if resp.StatusCode >= 500 {
		apiErr.Errors = []APIErrorDetail{{Message: fmt.Sprintf("Server error (%d)", resp.StatusCode)}}
		return apiErr
	}

	var errBody struct {
		Errors []APIErrorDetail `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&errBody); err == nil {
		apiErr.Errors = errBody.Errors
	}
	if len(apiErr.Errors) == 0 {
		apiErr.Errors = []APIErrorDetail{{Message: "Request failed"}}
	}
	return apiErr
}

func (c *Client) debugLog(method, path string, status int, duration time.Duration) {
	if c.debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] %s %s → %d (%s)\n", method, path, status, duration.Round(time.Millisecond))
	}
}
```

- [ ] **Step 5: Run tests**

```bash
cd /home/manu/Code/bitbucket-cli && go test ./internal/client/ -v -run TestClient
# Expected: all PASS
```

- [ ] **Step 6: Commit**

```bash
git add internal/client/client.go internal/client/errors.go internal/client/client_test.go
git commit -m "feat: add HTTP client with retry, debug logging, and error handling"
```

---

## Task 5: Pagination

**Files:**
- Create: `internal/client/pagination.go`, `internal/client/pagination_test.go`

- [ ] **Step 1: Write pagination tests**

Create `internal/client/pagination_test.go`:

```go
package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPaged(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := r.URL.Query().Get("start")
		w.Header().Set("Content-Type", "application/json")
		if start == "" || start == "0" {
			json.NewEncoder(w).Encode(map[string]any{
				"values":        []map[string]any{{"key": "PROJ1"}, {"key": "PROJ2"}},
				"size":          2,
				"start":         0,
				"limit":         2,
				"isLastPage":    false,
				"nextPageStart": 2,
			})
		} else {
			json.NewEncoder(w).Encode(map[string]any{
				"values":     []map[string]any{{"key": "PROJ3"}},
				"size":       1,
				"start":      2,
				"limit":      2,
				"isLastPage": true,
			})
		}
	}))
	defer server.Close()

	c := New(server.URL, "test-token")
	ctx := context.Background()

	// Test single page
	page, err := GetPaged[Project](ctx, c, "/projects", nil, 0, 2)
	if err != nil {
		t.Fatalf("GetPaged error: %v", err)
	}
	if len(page.Values) != 2 {
		t.Errorf("expected 2 values, got %d", len(page.Values))
	}
	if page.IsLastPage {
		t.Error("expected IsLastPage=false")
	}

	// Test auto-pagination
	all, err := GetAll[Project](ctx, c, "/projects", nil, 2)
	if err != nil {
		t.Fatalf("GetAll error: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 values, got %d", len(all))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /home/manu/Code/bitbucket-cli && go test ./internal/client/ -v -run TestGetPaged
# Expected: FAIL
```

- [ ] **Step 3: Implement pagination.go**

Create `internal/client/pagination.go`:

```go
package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/manu/bb/internal/validation"
)

// PagedResponse represents a paginated API response.
type PagedResponse[T any] struct {
	Values        []T  `json:"values"`
	Size          int  `json:"size"`
	Start         int  `json:"start"`
	Limit         int  `json:"limit"`
	IsLastPage    bool `json:"isLastPage"`
	NextPageStart int  `json:"nextPageStart"`
}

// GetPaged fetches a single page of results.
func GetPaged[T any](ctx context.Context, c *Client, path string, params url.Values, start, limit int) (*PagedResponse[T], error) {
	if params == nil {
		params = url.Values{}
	}
	params.Set("start", fmt.Sprintf("%d", validation.ClampStart(start)))
	params.Set("limit", fmt.Sprintf("%d", validation.ClampLimit(limit)))

	var result PagedResponse[T]
	if err := c.Get(ctx, path, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAll fetches all pages by following nextPageStart until isLastPage is true.
func GetAll[T any](ctx context.Context, c *Client, path string, params url.Values, limit int) ([]T, error) {
	var all []T
	start := 0
	clampedLimit := validation.ClampLimit(limit)

	for {
		page, err := GetPaged[T](ctx, c, path, params, start, clampedLimit)
		if err != nil {
			return nil, err
		}
		all = append(all, page.Values...)
		if page.IsLastPage {
			break
		}
		start = page.NextPageStart
	}
	return all, nil
}
```

- [ ] **Step 4: Run tests**

```bash
cd /home/manu/Code/bitbucket-cli && go test ./internal/client/ -v -run TestGetPaged
# Expected: PASS
```

- [ ] **Step 5: Commit**

```bash
git add internal/client/pagination.go internal/client/pagination_test.go
git commit -m "feat: add generic pagination with GetPaged and GetAll"
```

---

## Task 6: Output Formatters

**Files:**
- Create: `internal/output/formatter.go`, `internal/output/json.go`, `internal/output/table.go`, `internal/output/template.go`, `internal/output/formatter_test.go`

- [ ] **Step 1: Write formatter tests**

Create `internal/output/formatter_test.go`:

```go
package output

import (
	"strings"
	"testing"
)

type testItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestJSONFormatter(t *testing.T) {
	f := &JSONFormatter{}
	items := []testItem{{ID: 1, Name: "foo"}, {ID: 2, Name: "bar"}}
	out, err := f.Format(items)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if !strings.Contains(out, `"id": 1`) {
		t.Errorf("expected JSON with id:1, got: %s", out)
	}
}

func TestTemplateFormatter(t *testing.T) {
	f := &TemplateFormatter{Template: "{{range .}}{{.ID}}\t{{.Name}}\n{{end}}"}
	items := []testItem{{ID: 1, Name: "foo"}, {ID: 2, Name: "bar"}}
	out, err := f.Format(items)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if !strings.Contains(out, "1\tfoo") {
		t.Errorf("expected template output, got: %s", out)
	}
}

func TestTableFormatter(t *testing.T) {
	cols := []Column{
		{Header: "ID", Width: 5},
		{Header: "NAME", Width: 10},
	}
	f := NewTableFormatter(cols, false)
	rows := [][]string{{"1", "foo"}, {"2", "bar"}}
	out := f.FormatRows(rows)
	if !strings.Contains(out, "ID") {
		t.Errorf("expected header, got: %s", out)
	}
	if !strings.Contains(out, "foo") {
		t.Errorf("expected data, got: %s", out)
	}
}

func TestNewFormatter(t *testing.T) {
	f := NewFormatter(true, "", false)
	if _, ok := f.(*JSONFormatter); !ok {
		t.Error("expected JSONFormatter when json=true")
	}

	f = NewFormatter(false, "{{.ID}}", false)
	if _, ok := f.(*TemplateFormatter); !ok {
		t.Error("expected TemplateFormatter when format is set")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /home/manu/Code/bitbucket-cli && go test ./internal/output/ -v
# Expected: FAIL
```

- [ ] **Step 3: Implement formatter.go**

Create `internal/output/formatter.go`:

```go
package output

// Formatter formats data for output.
type Formatter interface {
	Format(data any) (string, error)
}

// Column defines a table column.
type Column struct {
	Header string
	Width  int
}

// NewFormatter creates the appropriate formatter based on flags.
func NewFormatter(jsonMode bool, format string, noColor bool) Formatter {
	if jsonMode {
		return &JSONFormatter{}
	}
	if format != "" {
		return &TemplateFormatter{Template: format}
	}
	return nil // Table formatting is done per-command
}
```

- [ ] **Step 4: Implement json.go**

Create `internal/output/json.go`:

```go
package output

import (
	"encoding/json"
)

type JSONFormatter struct{}

func (f *JSONFormatter) Format(data any) (string, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
```

- [ ] **Step 5: Implement template.go**

Create `internal/output/template.go`:

```go
package output

import (
	"bytes"
	"text/template"
)

type TemplateFormatter struct {
	Template string
}

func (f *TemplateFormatter) Format(data any) (string, error) {
	tmpl, err := template.New("output").Parse(f.Template)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
```

- [ ] **Step 6: Implement table.go**

Create `internal/output/table.go`:

```go
package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type TableFormatter struct {
	columns []Column
	noColor bool
}

func NewTableFormatter(columns []Column, noColor bool) *TableFormatter {
	return &TableFormatter{columns: columns, noColor: noColor}
}

func (f *TableFormatter) FormatRows(rows [][]string) string {
	if len(rows) == 0 {
		return "No results found."
	}

	isTTY := term.IsTerminal(int(os.Stdout.Fd()))
	useColor := isTTY && !f.noColor

	var sb strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true)
	for i, col := range f.columns {
		header := col.Header
		if useColor {
			header = headerStyle.Render(header)
		}
		if i < len(f.columns)-1 {
			sb.WriteString(fmt.Sprintf("%-*s  ", col.Width, header))
		} else {
			sb.WriteString(header)
		}
	}
	sb.WriteString("\n")

	// Rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(f.columns)-1 {
				truncated := truncate(cell, f.columns[i].Width)
				sb.WriteString(fmt.Sprintf("%-*s  ", f.columns[i].Width, truncated))
			} else {
				sb.WriteString(cell)
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// IsTerminal reports whether stdout is a terminal.
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
```

- [ ] **Step 7: Run tests**

```bash
cd /home/manu/Code/bitbucket-cli && go test ./internal/output/ -v
# Expected: PASS
```

- [ ] **Step 8: Commit**

```bash
git add internal/output/
git commit -m "feat: add output formatters (JSON, table, Go template)"
```

---

## Task 7: Config & Auth

**Files:**
- Create: `internal/config/config.go`, `internal/config/auth.go`, `internal/config/config_test.go`, `internal/config/auth_test.go`

- [ ] **Step 1: Write config tests**

Create `internal/config/config_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_EnvVarsOverride(t *testing.T) {
	t.Setenv("BITBUCKET_URL", "https://env.example.com")
	t.Setenv("BITBUCKET_TOKEN", "env-token")

	cfg, err := Load("default", "")
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.URL != "https://env.example.com" {
		t.Errorf("expected env URL, got %s", cfg.URL)
	}
	if cfg.Token != "env-token" {
		t.Errorf("expected env token, got %s", cfg.Token)
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	credsPath := filepath.Join(dir, "credentials.yaml")

	os.WriteFile(configPath, []byte(`
current-profile: default
profiles:
  default:
    url: https://file.example.com
    default-project: PROJ
    default-repo: my-repo
`), 0644)

	os.WriteFile(credsPath, []byte(`
profiles:
  default:
    token: file-token
`), 0600)

	cfg, err := LoadFromDir(dir, "default")
	if err != nil {
		t.Fatalf("LoadFromDir error: %v", err)
	}
	if cfg.URL != "https://file.example.com" {
		t.Errorf("expected file URL, got %s", cfg.URL)
	}
	if cfg.Token != "file-token" {
		t.Errorf("expected file token, got %s", cfg.Token)
	}
	if cfg.DefaultProject != "PROJ" {
		t.Errorf("expected default project PROJ, got %s", cfg.DefaultProject)
	}
}
```

- [ ] **Step 2: Write auth tests**

Create `internal/config/auth_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadCredentials(t *testing.T) {
	dir := t.TempDir()
	credsPath := filepath.Join(dir, "credentials.yaml")

	err := SaveCredentials(credsPath, "test-profile", "my-token")
	if err != nil {
		t.Fatalf("SaveCredentials error: %v", err)
	}

	// Verify file permissions
	info, err := os.Stat(credsPath)
	if err != nil {
		t.Fatalf("Stat error: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600 permissions, got %o", info.Mode().Perm())
	}

	token, err := LoadToken(credsPath, "test-profile")
	if err != nil {
		t.Fatalf("LoadToken error: %v", err)
	}
	if token != "my-token" {
		t.Errorf("expected my-token, got %s", token)
	}
}

func TestRemoveCredentials(t *testing.T) {
	dir := t.TempDir()
	credsPath := filepath.Join(dir, "credentials.yaml")

	SaveCredentials(credsPath, "test-profile", "my-token")
	err := RemoveCredentials(credsPath, "test-profile")
	if err != nil {
		t.Fatalf("RemoveCredentials error: %v", err)
	}

	token, err := LoadToken(credsPath, "test-profile")
	if err != nil {
		t.Fatalf("LoadToken error: %v", err)
	}
	if token != "" {
		t.Errorf("expected empty token after remove, got %s", token)
	}
}
```

- [ ] **Step 3: Implement config.go**

Create `internal/config/config.go`:

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the resolved configuration for a CLI session.
type Config struct {
	URL            string
	Token          string
	Profile        string
	DefaultProject string
	DefaultRepo    string
}

// ConfigDir returns the config directory path.
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "bb")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "bb")
}

// Load resolves configuration using: CLI flags > env vars > config files.
func Load(profile, configDir string) (*Config, error) {
	if configDir == "" {
		configDir = ConfigDir()
	}
	return LoadFromDir(configDir, profile)
}

// LoadFromDir loads config from a specific directory (useful for testing).
func LoadFromDir(dir, profile string) (*Config, error) {
	cfg := &Config{Profile: profile}

	// Load config file
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(dir)
	v.ReadInConfig() // Ignore error — file may not exist

	profilePrefix := fmt.Sprintf("profiles.%s", profile)

	cfg.URL = v.GetString(profilePrefix + ".url")
	cfg.DefaultProject = v.GetString(profilePrefix + ".default-project")
	cfg.DefaultRepo = v.GetString(profilePrefix + ".default-repo")

	// Load credentials
	credsPath := filepath.Join(dir, "credentials.yaml")
	token, _ := LoadToken(credsPath, profile)
	cfg.Token = token

	// Environment variables override config file
	if envURL := os.Getenv("BITBUCKET_URL"); envURL != "" {
		cfg.URL = envURL
	}
	if envToken := os.Getenv("BITBUCKET_TOKEN"); envToken != "" {
		cfg.Token = envToken
	}

	return cfg, nil
}

// SaveProfile writes a profile to the config file.
func SaveProfile(configDir, profile, url, defaultProject, defaultRepo string) error {
	configPath := filepath.Join(configDir, "config.yaml")
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	v.ReadInConfig() // Load existing

	prefix := fmt.Sprintf("profiles.%s", profile)
	v.Set(prefix+".url", url)
	if defaultProject != "" {
		v.Set(prefix+".default-project", defaultProject)
	}
	if defaultRepo != "" {
		v.Set(prefix+".default-repo", defaultRepo)
	}
	v.Set("current-profile", profile)

	os.MkdirAll(configDir, 0755)
	return v.WriteConfigAs(configPath)
}

// ClearDefaults removes default-project and default-repo from a profile.
func ClearDefaults(configDir, profile string) error {
	configPath := filepath.Join(configDir, "config.yaml")
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	v.ReadInConfig()

	prefix := fmt.Sprintf("profiles.%s", profile)
	v.Set(prefix+".default-project", "")
	v.Set(prefix+".default-repo", "")

	return v.WriteConfigAs(configPath)
}
```

- [ ] **Step 4: Implement auth.go**

Create `internal/config/auth.go`:

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// SaveCredentials stores a token for a profile in credentials.yaml with 0600 permissions.
func SaveCredentials(credsPath, profile, token string) error {
	os.MkdirAll(filepath.Dir(credsPath), 0755)

	v := viper.New()
	v.SetConfigFile(credsPath)
	v.SetConfigType("yaml")
	v.ReadInConfig() // Load existing

	v.Set(fmt.Sprintf("profiles.%s.token", profile), token)

	if err := v.WriteConfigAs(credsPath); err != nil {
		return err
	}
	return os.Chmod(credsPath, 0600)
}

// LoadToken reads a token for a profile from credentials.yaml.
func LoadToken(credsPath, profile string) (string, error) {
	v := viper.New()
	v.SetConfigFile(credsPath)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return "", nil // No file = no token
	}
	return v.GetString(fmt.Sprintf("profiles.%s.token", profile)), nil
}

// RemoveCredentials removes a profile's token from credentials.yaml.
func RemoveCredentials(credsPath, profile string) error {
	v := viper.New()
	v.SetConfigFile(credsPath)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil
	}

	v.Set(fmt.Sprintf("profiles.%s.token", profile), "")
	return v.WriteConfigAs(credsPath)
}
```

- [ ] **Step 5: Run tests**

```bash
cd /home/manu/Code/bitbucket-cli && go test ./internal/config/ -v
# Expected: PASS
```

- [ ] **Step 6: Commit**

```bash
git add internal/config/
git commit -m "feat: add config and auth management with profile support"
```

---

## Task 8: Confirmation Helpers

**Files:**
- Create: `internal/cmd/confirm.go`

- [ ] **Step 1: Create confirm.go**

```go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfirmDangerous prompts the user to type the resource name to confirm deletion.
// Returns true if confirmed, false otherwise. Skips prompt if confirmed=true.
func ConfirmDangerous(resourceType, resourceName string, confirmed bool) bool {
	if confirmed {
		return true
	}

	fmt.Fprintf(os.Stderr, "\n⚠ This will permanently delete %s '%s'. This cannot be undone.\n", resourceType, resourceName)
	fmt.Fprintf(os.Stderr, "Type the %s name to confirm: ", resourceType)

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()) == resourceName
	}
	return false
}

// ConfirmDestructive prompts for destructive operations (project/repo delete).
// Requires both --confirm and --i-understand-this-is-destructive for scripting.
func ConfirmDestructive(resourceType, resourceName string, confirmed, understands bool) bool {
	if confirmed && understands {
		return true
	}

	fmt.Fprintf(os.Stderr, "\n🛑 DESTRUCTIVE: This will permanently delete %s '%s' and ALL its contents.\n", resourceType, resourceName)
	fmt.Fprintf(os.Stderr, "   This CANNOT be undone.\n")
	fmt.Fprintf(os.Stderr, "Type '%s' to confirm: ", resourceName)

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()) == resourceName
	}
	return false
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/cmd/confirm.go
git commit -m "feat: add confirmation helpers for dangerous and destructive operations"
```

---

## Task 9: Client API Methods — All Resources

**Files:**
- Create: `internal/client/projects.go`, `internal/client/repositories.go`, `internal/client/branches.go`, `internal/client/pull_requests.go`, `internal/client/pull_request_comments.go`, `internal/client/pull_request_tasks.go`, `internal/client/commits.go`, `internal/client/files.go`, `internal/client/search.go`, `internal/client/users.go`, `internal/client/dashboard.go`, `internal/client/attachments.go`, `internal/client/dangerous.go`

This task creates all client API methods. Each method maps directly to a Bitbucket REST API endpoint. The pattern is consistent: validate inputs, build path/params/body, call client.Get/Post/Put/Delete, return typed result.

**API path reference (from Python MCP server):**
- Projects: `GET /projects`, `GET /projects/{key}`
- Repos: `GET /projects/{key}/repos`, `GET /projects/{key}/repos/{slug}`, `POST /projects/{key}/repos`
- Branches: `GET .../branches`, `GET .../branches/default`, `POST .../branches`
- Tags: `GET .../tags`
- PRs: `GET/POST/PUT .../pull-requests`, `POST .../pull-requests/{id}/merge`, `POST .../pull-requests/{id}/decline`, `POST .../pull-requests/{id}/reopen`, `POST .../pull-requests/{id}/approve`, `DELETE .../pull-requests/{id}/approve`, `PUT .../pull-requests/{id}/participants/status`
- PR comments: `GET .../comments`, `GET .../comments/{id}`, `POST .../comments`, `PUT .../comments/{id}`
- PR tasks: `GET .../tasks`, `POST .../tasks`, `GET .../tasks/{id}`, `PUT .../tasks/{id}`
- Commits: `GET .../commits`, `GET .../commits/{id}`, `GET .../commits/{id}/diff`, `GET .../commits/{id}/changes`
- Files: `GET .../browse`, `GET .../raw/{path}`, `GET .../files`
- Search: `POST /rest/search/latest/search` (fallback GET)
- Users: `GET /users?filter=`
- Dashboard: `GET /dashboard/pull-requests`, `GET /inbox/pull-requests`
- Attachments: `GET .../attachments/{id}` (raw), `GET .../attachments/{id}/metadata`, `PUT .../attachments/{id}/metadata`
- Dangerous: `POST /rest/branch-utils/1.0/.../branches`, `DELETE /rest/git/1.0/.../tags/{name}`, `DELETE .../pull-requests/{id}`, `DELETE .../comments/{id}`, `DELETE .../tasks/{id}`, `DELETE .../attachments/{id}`, `DELETE .../attachments/{id}/metadata`, `DELETE /projects/{key}`, `DELETE /projects/{key}/repos/{slug}`

- [ ] **Step 1: Create projects.go**

```go
package client

import (
	"context"
	"net/url"
)

func (c *Client) ListProjects(ctx context.Context, start, limit int) (*PagedResponse[Project], error) {
	return GetPaged[Project](ctx, c, "/projects", nil, start, limit)
}

func (c *Client) GetProject(ctx context.Context, key string) (*Project, error) {
	var result Project
	err := c.Get(ctx, "/projects/"+key, nil, &result)
	return &result, err
}
```

- [ ] **Step 2: Create repositories.go**

```go
package client

import (
	"context"
	"net/url"
)

func repoPath(project, repo string) string {
	return "/projects/" + project + "/repos/" + repo
}

func (c *Client) ListRepositories(ctx context.Context, project string, start, limit int) (*PagedResponse[Repository], error) {
	return GetPaged[Repository](ctx, c, "/projects/"+project+"/repos", nil, start, limit)
}

func (c *Client) GetRepository(ctx context.Context, project, repo string) (*Repository, error) {
	var result Repository
	err := c.Get(ctx, repoPath(project, repo), nil, &result)
	return &result, err
}

func (c *Client) CreateRepository(ctx context.Context, project, name, description string, forkable bool) (*Repository, error) {
	body := map[string]any{
		"name":     name,
		"scmId":    "git",
		"forkable": forkable,
	}
	if description != "" {
		body["description"] = description
	}
	var result Repository
	err := c.Post(ctx, "/projects/"+project+"/repos", body, nil, &result)
	return &result, err
}
```

- [ ] **Step 3: Create branches.go**

```go
package client

import (
	"context"
	"net/url"
)

func (c *Client) ListBranches(ctx context.Context, project, repo, filter string, start, limit int) (*PagedResponse[Branch], error) {
	params := url.Values{}
	if filter != "" {
		params.Set("filterText", filter)
	}
	return GetPaged[Branch](ctx, c, repoPath(project, repo)+"/branches", params, start, limit)
}

func (c *Client) GetDefaultBranch(ctx context.Context, project, repo string) (*Branch, error) {
	var result Branch
	err := c.Get(ctx, repoPath(project, repo)+"/branches/default", nil, &result)
	return &result, err
}

func (c *Client) CreateBranch(ctx context.Context, project, repo, name, startPoint string) (*Branch, error) {
	body := map[string]any{"name": name, "startPoint": startPoint}
	var result Branch
	err := c.Post(ctx, repoPath(project, repo)+"/branches", body, nil, &result)
	return &result, err
}

func (c *Client) ListTags(ctx context.Context, project, repo, filter string, start, limit int) (*PagedResponse[Tag], error) {
	params := url.Values{}
	if filter != "" {
		params.Set("filterText", filter)
	}
	return GetPaged[Tag](ctx, c, repoPath(project, repo)+"/tags", params, start, limit)
}
```

- [ ] **Step 4: Create pull_requests.go**

This is the largest file. Contains all PR CRUD, merge/decline/reopen, approval, watching, diff, and participants methods.

```go
package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

func prPath(project, repo string, id int) string {
	return fmt.Sprintf("%s/pull-requests/%d", repoPath(project, repo), id)
}

func (c *Client) ListPullRequests(ctx context.Context, project, repo string, opts ListPROptions) (*PagedResponse[PullRequest], error) {
	params := url.Values{}
	if opts.State != "" {
		params.Set("state", strings.ToUpper(opts.State))
	}
	if opts.Direction != "" {
		params.Set("direction", strings.ToUpper(opts.Direction))
	}
	if opts.Order != "" {
		params.Set("order", strings.ToUpper(opts.Order))
	}
	if opts.At != "" {
		params.Set("at", opts.At)
	}
	if opts.FilterText != "" {
		params.Set("filterText", opts.FilterText)
	}
	if opts.Participant != "" {
		params.Set("role.1", "PARTICIPANT")
		params.Set("username.1", opts.Participant)
	}
	if opts.Draft != nil {
		params.Set("draft", fmt.Sprintf("%t", *opts.Draft))
	}
	return GetPaged[PullRequest](ctx, c, repoPath(project, repo)+"/pull-requests", params, opts.Start, opts.Limit)
}

// ListPROptions holds filter options for listing pull requests.
type ListPROptions struct {
	State       string
	Direction   string
	Order       string
	At          string
	FilterText  string
	Participant string
	Draft       *bool
	Start       int
	Limit       int
}

func (c *Client) GetPullRequest(ctx context.Context, project, repo string, id int) (*PullRequest, error) {
	var result PullRequest
	err := c.Get(ctx, prPath(project, repo, id), nil, &result)
	return &result, err
}

func (c *Client) CreatePullRequest(ctx context.Context, project, repo string, input CreatePRInput) (*PullRequest, error) {
	fromRef := input.FromRef
	if !strings.HasPrefix(fromRef, "refs/") {
		fromRef = "refs/heads/" + fromRef
	}
	toRef := input.ToRef
	if !strings.HasPrefix(toRef, "refs/") {
		toRef = "refs/heads/" + toRef
	}

	body := map[string]any{
		"title":       input.Title,
		"description": input.Description,
		"fromRef": map[string]any{
			"id":         fromRef,
			"repository": map[string]any{"slug": repo, "project": map[string]any{"key": project}},
		},
		"toRef": map[string]any{
			"id":         toRef,
			"repository": map[string]any{"slug": repo, "project": map[string]any{"key": project}},
		},
	}
	if len(input.Reviewers) > 0 {
		reviewers := make([]map[string]any, len(input.Reviewers))
		for i, r := range input.Reviewers {
			reviewers[i] = map[string]any{"user": map[string]any{"name": r}}
		}
		body["reviewers"] = reviewers
	}
	if input.Draft {
		body["draft"] = true
	}

	var result PullRequest
	err := c.Post(ctx, repoPath(project, repo)+"/pull-requests", body, nil, &result)
	return &result, err
}

type CreatePRInput struct {
	Title       string
	Description string
	FromRef     string
	ToRef       string
	Reviewers   []string
	Draft       bool
}

func (c *Client) UpdatePullRequest(ctx context.Context, project, repo string, id, version int, input UpdatePRInput) (*PullRequest, error) {
	// Fetch current PR to preserve unchanged fields
	current, err := c.GetPullRequest(ctx, project, repo, id)
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"version":     version,
		"title":       current.Title,
		"description": current.Description,
		"toRef":       map[string]any{"id": current.ToRef.ID, "repository": map[string]any{"slug": repo, "project": map[string]any{"key": project}}},
		"reviewers":   current.Reviewers,
	}
	if input.Title != "" {
		body["title"] = input.Title
	}
	if input.Description != nil {
		body["description"] = *input.Description
	}
	if input.Reviewers != nil {
		reviewers := make([]map[string]any, len(input.Reviewers))
		for i, r := range input.Reviewers {
			reviewers[i] = map[string]any{"user": map[string]any{"name": r}}
		}
		body["reviewers"] = reviewers
	}
	if input.TargetBranch != "" {
		toRef := input.TargetBranch
		if !strings.HasPrefix(toRef, "refs/") {
			toRef = "refs/heads/" + toRef
		}
		body["toRef"] = map[string]any{
			"id":         toRef,
			"repository": map[string]any{"slug": repo, "project": map[string]any{"key": project}},
		}
	}
	if input.Draft != nil {
		body["draft"] = *input.Draft
	}

	var result PullRequest
	err = c.Put(ctx, prPath(project, repo, id), body, nil, &result)
	return &result, err
}

type UpdatePRInput struct {
	Title        string
	Description  *string
	Reviewers    []string
	TargetBranch string
	Draft        *bool
}

func (c *Client) MergePullRequest(ctx context.Context, project, repo string, id, version int, strategy string) (*PullRequest, error) {
	params := url.Values{"version": {fmt.Sprintf("%d", version)}}
	if strategy != "" {
		params.Set("strategyId", strategy)
	}
	var result PullRequest
	err := c.Post(ctx, prPath(project, repo, id)+"/merge", nil, params, &result)
	return &result, err
}

func (c *Client) DeclinePullRequest(ctx context.Context, project, repo string, id, version int) (*PullRequest, error) {
	params := url.Values{"version": {fmt.Sprintf("%d", version)}}
	var result PullRequest
	err := c.Post(ctx, prPath(project, repo, id)+"/decline", nil, params, &result)
	return &result, err
}

func (c *Client) ReopenPullRequest(ctx context.Context, project, repo string, id, version int) (*PullRequest, error) {
	params := url.Values{"version": {fmt.Sprintf("%d", version)}}
	var result PullRequest
	err := c.Post(ctx, prPath(project, repo, id)+"/reopen", nil, params, &result)
	return &result, err
}

func (c *Client) ApprovePullRequest(ctx context.Context, project, repo string, id int) (*Participant, error) {
	var result Participant
	err := c.Post(ctx, prPath(project, repo, id)+"/approve", nil, nil, &result)
	return &result, err
}

func (c *Client) UnapprovePullRequest(ctx context.Context, project, repo string, id int) error {
	return c.Delete(ctx, prPath(project, repo, id)+"/approve", nil, nil)
}

func (c *Client) RequestChanges(ctx context.Context, project, repo string, id int) (*Participant, error) {
	var result Participant
	err := c.Put(ctx, prPath(project, repo, id)+"/participants/status", map[string]any{"status": "NEEDS_WORK"}, nil, &result)
	return &result, err
}

func (c *Client) RemoveChangeRequest(ctx context.Context, project, repo string, id int) (*Participant, error) {
	var result Participant
	err := c.Put(ctx, prPath(project, repo, id)+"/participants/status", map[string]any{"status": "UNAPPROVED"}, nil, &result)
	return &result, err
}

func (c *Client) CanMerge(ctx context.Context, project, repo string, id int) (*MergeStatus, error) {
	var result MergeStatus
	err := c.Get(ctx, prPath(project, repo, id)+"/merge", nil, &result)
	return &result, err
}

func (c *Client) WatchPullRequest(ctx context.Context, project, repo string, id int) error {
	return c.Post(ctx, prPath(project, repo, id)+"/watch", nil, nil, nil)
}

func (c *Client) UnwatchPullRequest(ctx context.Context, project, repo string, id int) error {
	return c.Delete(ctx, prPath(project, repo, id)+"/watch", nil, nil)
}

func (c *Client) GetCommitMessageSuggestion(ctx context.Context, project, repo string, id int) (map[string]any, error) {
	var result map[string]any
	err := c.Get(ctx, prPath(project, repo, id)+"/commit-message-suggestion", nil, &result)
	return result, err
}

func (c *Client) GetPullRequestDiff(ctx context.Context, project, repo string, id, contextLines int, srcPath string) (*Diff, error) {
	params := url.Values{"contextLines": {fmt.Sprintf("%d", contextLines)}}
	if srcPath != "" {
		params.Set("srcPath", srcPath)
	}
	var result Diff
	err := c.Get(ctx, prPath(project, repo, id)+"/diff", params, &result)
	return &result, err
}

func (c *Client) GetPullRequestDiffStat(ctx context.Context, project, repo string, id, start, limit int) (*PagedResponse[Change], error) {
	return GetPaged[Change](ctx, c, prPath(project, repo, id)+"/changes", nil, start, limit)
}

func (c *Client) ListPullRequestCommits(ctx context.Context, project, repo string, id, start, limit int) (*PagedResponse[Commit], error) {
	return GetPaged[Commit](ctx, c, prPath(project, repo, id)+"/commits", nil, start, limit)
}

func (c *Client) GetPullRequestActivities(ctx context.Context, project, repo string, id, start, limit int) (*PagedResponse[Activity], error) {
	return GetPaged[Activity](ctx, c, prPath(project, repo, id)+"/activities", nil, start, limit)
}

func (c *Client) ListPullRequestParticipants(ctx context.Context, project, repo string, id, start, limit int) (*PagedResponse[Participant], error) {
	return GetPaged[Participant](ctx, c, prPath(project, repo, id)+"/participants", nil, start, limit)
}

func (c *Client) PublishDraft(ctx context.Context, project, repo string, id, version int) (*PullRequest, error) {
	current, err := c.GetPullRequest(ctx, project, repo, id)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"version":     version,
		"title":       current.Title,
		"description": current.Description,
		"draft":       false,
		"toRef":       map[string]any{"id": current.ToRef.ID, "repository": map[string]any{"slug": repo, "project": map[string]any{"key": project}}},
		"reviewers":   current.Reviewers,
	}
	var result PullRequest
	err = c.Put(ctx, prPath(project, repo, id), body, nil, &result)
	return &result, err
}

func (c *Client) ConvertToDraft(ctx context.Context, project, repo string, id, version int) (*PullRequest, error) {
	current, err := c.GetPullRequest(ctx, project, repo, id)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"version":     version,
		"title":       current.Title,
		"description": current.Description,
		"draft":       true,
		"toRef":       map[string]any{"id": current.ToRef.ID, "repository": map[string]any{"slug": repo, "project": map[string]any{"key": project}}},
		"reviewers":   current.Reviewers,
	}
	var result PullRequest
	err = c.Put(ctx, prPath(project, repo, id), body, nil, &result)
	return &result, err
}
```

- [ ] **Step 5: Create pull_request_comments.go**

```go
package client

import (
	"context"
	"fmt"
	"net/url"
)

func (c *Client) ListPRComments(ctx context.Context, project, repo string, prID, start, limit int) (*PagedResponse[Comment], error) {
	return GetPaged[Comment](ctx, c, prPath(project, repo, prID)+"/comments", nil, start, limit)
}

func (c *Client) GetPRComment(ctx context.Context, project, repo string, prID, commentID int) (*Comment, error) {
	var result Comment
	err := c.Get(ctx, fmt.Sprintf("%s/comments/%d", prPath(project, repo, prID), commentID), nil, &result)
	return &result, err
}

func (c *Client) AddPRComment(ctx context.Context, project, repo string, prID int, input AddCommentInput) (*Comment, error) {
	body := map[string]any{"text": input.Text}
	if input.Severity != "" {
		body["severity"] = input.Severity
	}
	if input.ParentID != nil {
		body["parent"] = map[string]any{"id": *input.ParentID}
	}
	if input.FilePath != "" {
		anchor := map[string]any{"path": input.FilePath}
		if input.Line != nil {
			anchor["line"] = *input.Line
		}
		if input.LineType != "" {
			anchor["lineType"] = input.LineType
		}
		if input.FileType != "" {
			anchor["fileType"] = input.FileType
		}
		body["anchor"] = anchor
	}
	var result Comment
	err := c.Post(ctx, prPath(project, repo, prID)+"/comments", body, nil, &result)
	return &result, err
}

type AddCommentInput struct {
	Text     string
	Severity string
	ParentID *int
	FilePath string
	Line     *int
	LineType string
	FileType string
}

func (c *Client) UpdatePRComment(ctx context.Context, project, repo string, prID, commentID, version int, text string) (*Comment, error) {
	var result Comment
	err := c.Put(ctx, fmt.Sprintf("%s/comments/%d", prPath(project, repo, prID), commentID), map[string]any{"text": text, "version": version}, nil, &result)
	return &result, err
}

func (c *Client) ResolvePRComment(ctx context.Context, project, repo string, prID, commentID, version int) (*Comment, error) {
	var result Comment
	err := c.Put(ctx, fmt.Sprintf("%s/comments/%d", prPath(project, repo, prID), commentID), map[string]any{"state": "RESOLVED", "version": version}, nil, &result)
	return &result, err
}

func (c *Client) ReopenPRComment(ctx context.Context, project, repo string, prID, commentID, version int) (*Comment, error) {
	var result Comment
	err := c.Put(ctx, fmt.Sprintf("%s/comments/%d", prPath(project, repo, prID), commentID), map[string]any{"state": "OPEN", "version": version}, nil, &result)
	return &result, err
}

func (c *Client) DeletePRComment(ctx context.Context, project, repo string, prID, commentID, version int) error {
	params := url.Values{"version": {fmt.Sprintf("%d", version)}}
	return c.Delete(ctx, fmt.Sprintf("%s/comments/%d", prPath(project, repo, prID), commentID), params, nil)
}
```

- [ ] **Step 6: Create pull_request_tasks.go**

```go
package client

import (
	"context"
	"fmt"
)

func (c *Client) ListPRTasks(ctx context.Context, project, repo string, prID, start, limit int) (*PagedResponse[Task], error) {
	return GetPaged[Task](ctx, c, prPath(project, repo, prID)+"/tasks", nil, start, limit)
}

func (c *Client) GetPRTask(ctx context.Context, project, repo string, prID, taskID int) (*Task, error) {
	var result Task
	err := c.Get(ctx, fmt.Sprintf("%s/tasks/%d", prPath(project, repo, prID), taskID), nil, &result)
	return &result, err
}

func (c *Client) CreatePRTask(ctx context.Context, project, repo string, prID int, text string, commentID *int) (*Task, error) {
	body := map[string]any{"text": text}
	if commentID != nil {
		body["anchor"] = map[string]any{"id": *commentID, "type": "COMMENT"}
	}
	var result Task
	err := c.Post(ctx, prPath(project, repo, prID)+"/tasks", body, nil, &result)
	return &result, err
}

func (c *Client) UpdatePRTask(ctx context.Context, project, repo string, prID, taskID int, text, state string) (*Task, error) {
	body := map[string]any{}
	if text != "" {
		body["text"] = text
	}
	if state != "" {
		body["state"] = state
	}
	var result Task
	err := c.Put(ctx, fmt.Sprintf("%s/tasks/%d", prPath(project, repo, prID), taskID), body, nil, &result)
	return &result, err
}

func (c *Client) DeletePRTask(ctx context.Context, project, repo string, prID, taskID int) error {
	return c.Delete(ctx, fmt.Sprintf("%s/tasks/%d", prPath(project, repo, prID), taskID), nil, nil)
}
```

- [ ] **Step 7: Create commits.go, files.go, search.go, users.go, dashboard.go, attachments.go, dangerous.go**

Each follows the same pattern. Create all remaining client files:

**commits.go** — `ListCommits`, `GetCommit`, `GetCommitDiff`, `GetCommitChanges`
**files.go** — `BrowseFiles`, `GetFileContent`, `ListFiles`
**search.go** — `SearchCode`, `FindFile` (using `c.Search()` with POST body)
**users.go** — `FindUser`
**dashboard.go** — `ListDashboardPRs`, `ListInboxPRs`
**attachments.go** — `GetAttachment`, `GetAttachmentMetadata`, `SaveAttachmentMetadata`
**dangerous.go** — `DeleteBranch` (POST `/rest/branch-utils/1.0/.../branches`), `DeleteTag` (DELETE `/rest/git/1.0/.../tags/{name}`), `DeletePullRequest`, `DeleteAttachment`, `DeleteAttachmentMetadata`, `DeleteProject`, `DeleteRepository`

Implement each following the exact API paths from the Python reference. See the Python tool files for exact path templates, HTTP methods, and body shapes.

- [ ] **Step 8: Verify all client code compiles**

```bash
cd /home/manu/Code/bitbucket-cli && go build ./internal/client/
# Expected: Success
```

- [ ] **Step 9: Commit**

```bash
git add internal/client/
git commit -m "feat: add all Bitbucket API client methods for 66 operations"
```

---

## Task 10: Cobra Commands — All Command Groups

**Files:**
- Create: all files in `internal/cmd/` (auth.go, project.go, repo.go, branch.go, tag.go, pr.go, pr_comment.go, pr_task.go, commit.go, file.go, search.go, user.go, dashboard.go, attachment.go)
- Modify: `internal/cmd/root.go` (wire subcommands)

Each command follows this pattern:
1. Create parent command (e.g., `bb project`)
2. Add subcommands (e.g., `bb project list`, `bb project get`)
3. Each subcommand's RunE: resolve config → create client → validate args → call client method → format output
4. Handle `--json`/`--format` flags for output formatting
5. Handle `--limit`/`--page`/`--all` for pagination
6. Handle `--confirm`/`--i-understand-this-is-destructive` for dangerous operations

- [ ] **Step 1: Update root.go to wire everything**

Update `internal/cmd/root.go` to initialize config, create client, and add all subcommands:

```go
// Add to NewRootCmd:
rootCmd.AddCommand(NewAuthCmd(flags))
rootCmd.AddCommand(NewProjectCmd(flags))
rootCmd.AddCommand(NewRepoCmd(flags))
rootCmd.AddCommand(NewBranchCmd(flags))
rootCmd.AddCommand(NewTagCmd(flags))
rootCmd.AddCommand(NewPRCmd(flags))
rootCmd.AddCommand(NewCommitCmd(flags))
rootCmd.AddCommand(NewFileCmd(flags))
rootCmd.AddCommand(NewSearchCmd(flags))
rootCmd.AddCommand(NewUserCmd(flags))
rootCmd.AddCommand(NewDashboardCmd(flags))
rootCmd.AddCommand(NewAttachmentCmd(flags))
```

Add a helper to create a client from config:

```go
func newClient(flags *GlobalFlags) (*client.Client, *config.Config, error) {
	cfg, err := config.Load(flags.Profile, "")
	if err != nil {
		return nil, nil, err
	}
	if cfg.URL == "" {
		return nil, nil, fmt.Errorf("no Bitbucket URL configured. Run 'bb auth login' or set BITBUCKET_URL")
	}
	if cfg.Token == "" {
		return nil, nil, fmt.Errorf("no token configured. Run 'bb auth login' or set BITBUCKET_TOKEN")
	}
	c := client.New(cfg.URL, cfg.Token)
	c.SetDebug(flags.Debug)
	return c, cfg, nil
}

func resolveProjectRepo(cfg *config.Config, args []string, minArgs int) (string, string, []string, error) {
	// If default project/repo are set and args are short, use defaults
	project := cfg.DefaultProject
	repo := cfg.DefaultRepo
	remaining := args

	if len(args) >= 2 {
		project = args[0]
		repo = args[1]
		remaining = args[2:]
	} else if len(args) == 1 && project != "" {
		repo = args[0]
		remaining = args[1:]
	}

	if project == "" {
		return "", "", nil, fmt.Errorf("project is required (provide as argument or set with 'bb repo use')")
	}
	if minArgs >= 2 && repo == "" {
		return "", "", nil, fmt.Errorf("repository is required (provide as argument or set with 'bb repo use')")
	}

	return project, repo, remaining, nil
}
```

- [ ] **Step 2: Create auth.go**

`bb auth login`, `bb auth logout`, `bb auth status` — interactive login, credential management.

- [ ] **Step 3: Create project.go**

`bb project list`, `bb project get <key>`, `bb project delete <key>` [destructive]

- [ ] **Step 4: Create repo.go**

`bb repo list`, `bb repo get`, `bb repo create`, `bb repo delete` [destructive], `bb repo use`, `bb repo clear`

- [ ] **Step 5: Create branch.go**

`bb branch list`, `bb branch create`, `bb branch default`, `bb branch delete` [dangerous]

- [ ] **Step 6: Create tag.go**

`bb tag list`, `bb tag delete` [dangerous]

- [ ] **Step 7: Create pr.go**

All 25 PR subcommands. This will be the largest command file. Each subcommand calls the corresponding client method.

- [ ] **Step 8: Create pr_comment.go**

`bb pr comment list/get/add/update/resolve/reopen/delete`

- [ ] **Step 9: Create pr_task.go**

`bb pr task list/create/get/update/delete`

- [ ] **Step 10: Create commit.go**

`bb commit list/get/diff/changes`

- [ ] **Step 11: Create file.go**

`bb file browse/cat/list/find`

- [ ] **Step 12: Create search.go**

`bb search code`

- [ ] **Step 13: Create user.go**

`bb user find`

- [ ] **Step 14: Create dashboard.go**

`bb dashboard list/inbox`

- [ ] **Step 15: Create attachment.go**

`bb attachment get/meta/save-meta/delete/delete-meta`

- [ ] **Step 16: Verify build**

```bash
cd /home/manu/Code/bitbucket-cli && make build
bin/bb --help
bin/bb pr --help
bin/bb pr list --help
# Expected: Help output with all commands and flags
```

- [ ] **Step 17: Commit**

```bash
git add internal/cmd/ cmd/
git commit -m "feat: add all 66 Cobra commands across 12 resource groups"
```

---

## Task 11: Integration Build & Smoke Test

**Files:**
- Verify: full build, all commands accessible

- [ ] **Step 1: Full build and binary size check**

```bash
cd /home/manu/Code/bitbucket-cli && make build
ls -lh bin/bb
# Expected: Single binary, ~10-15MB
```

- [ ] **Step 2: Verify all command help outputs**

```bash
bin/bb --help
bin/bb auth --help
bin/bb project --help
bin/bb repo --help
bin/bb branch --help
bin/bb tag --help
bin/bb pr --help
bin/bb pr comment --help
bin/bb pr task --help
bin/bb commit --help
bin/bb file --help
bin/bb search --help
bin/bb user --help
bin/bb dashboard --help
bin/bb attachment --help
bin/bb completion --help
```

- [ ] **Step 3: Run all tests**

```bash
cd /home/manu/Code/bitbucket-cli && make test
# Expected: All tests pass
```

- [ ] **Step 4: Generate shell completions**

```bash
bin/bb completion bash > /dev/null
bin/bb completion zsh > /dev/null
bin/bb completion fish > /dev/null
# Expected: No errors
```

- [ ] **Step 5: Commit any fixes**

```bash
git add -A
git commit -m "fix: integration fixes from smoke test"
```

---

## Task 12: Final Polish

**Files:**
- Create: `README.md`

- [ ] **Step 1: Create README.md**

Brief README with installation instructions, quick start, and command overview.

- [ ] **Step 2: Tag v0.1.0**

```bash
git tag v0.1.0
make build
bin/bb --version
# Expected: bb version v0.1.0
```

- [ ] **Step 3: Final commit**

```bash
git add README.md
git commit -m "docs: add README with installation and usage guide"
```
