# bb — Bitbucket Server CLI (Go)

**Date:** 2026-03-24
**Status:** Approved

## Overview

A Go CLI tool (`bb`) that wraps the Bitbucket Server REST API, providing full parity with the 66 operations exposed by the [bitbucket-server-mcp](../../../) Python MCP server. Git-style subcommands, multiple output formats, profile-based auth, and two-tier safety for dangerous operations.

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Repository | Separate repo (`bitbucket-cli`) | Clean separation from the Python MCP server |
| UX model | Git-style subcommands (`bb pr list`) | Matches `gh`, `glab` conventions developers expect |
| Scope | Full parity (all 66 operations) | Complete coverage from v1 |
| Auth | Config file + env vars + `bb auth login` | Best onboarding UX, script-friendly via env vars |
| Output | Table + JSON + Go templates | Covers human, machine, and custom formatting needs |
| Binary name | `bb` | Short, memorable, mirrors `gh`/`glab` convention |
| Dependencies | Cobra + Viper + Lipgloss | De facto standard Go CLI stack |
| Architecture | Flat client + command layer | Simple, testable, easy to navigate |

---

## 1. Project Structure

```
bb/
├── cmd/bb/
│   └── main.go                  # Entry point
├── internal/
│   ├── cmd/
│   │   ├── root.go              # Root command, global flags
│   │   ├── auth.go              # bb auth login / logout / status
│   │   ├── project.go           # bb project list / get
│   │   ├── repo.go              # bb repo list / get / create / use / clear
│   │   ├── branch.go            # bb branch list / create / default / delete
│   │   ├── tag.go               # bb tag list / delete
│   │   ├── pr.go                # bb pr list / get / create / update / merge / ...
│   │   ├── pr_comment.go        # bb pr comment list / add / update / resolve / ...
│   │   ├── pr_task.go           # bb pr task list / create / update / ...
│   │   ├── commit.go            # bb commit list / get / diff / changes
│   │   ├── file.go              # bb file browse / cat / list / find
│   │   ├── search.go            # bb search code
│   │   ├── user.go              # bb user find
│   │   ├── dashboard.go         # bb dashboard list / inbox
│   │   └── attachment.go        # bb attachment get / meta / ...
│   ├── client/
│   │   ├── client.go            # HTTP client struct, auth, error handling
│   │   ├── pagination.go        # Generic paginated request helper
│   │   ├── types.go             # All API data types
│   │   ├── projects.go          # API methods for projects
│   │   ├── repositories.go      # API methods for repos
│   │   ├── branches.go          # API methods for branches & tags
│   │   ├── pull_requests.go     # API methods for PRs
│   │   ├── commits.go           # API methods for commits
│   │   ├── files.go             # API methods for file browsing/content
│   │   ├── search.go            # API methods for code search
│   │   ├── users.go             # API methods for user lookup
│   │   ├── dashboard.go         # API methods for dashboard
│   │   └── attachments.go       # API methods for attachments
│   ├── config/
│   │   ├── config.go            # Viper-based config loading, profile support
│   │   └── auth.go              # Token storage, credentials file
│   ├── output/
│   │   ├── formatter.go         # Interface + factory: table, JSON, template
│   │   ├── table.go             # Table rendering (Lipgloss)
│   │   ├── json.go              # JSON output
│   │   └── template.go          # Go template rendering
│   └── validation/
│       └── validation.go        # Input validators
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 2. Command Tree

```
bb
├── auth
│   ├── login                    # Interactive or --url/--token
│   ├── logout                   # Remove credentials
│   └── status                   # Show auth state
│
├── project
│   ├── list                     # list_projects
│   └── get <key>                # get_project
│
├── repo
│   ├── list <project>           # list_repositories
│   ├── get <project> <slug>     # get_repository
│   ├── create <project> <name>  # create_repository (--description, --forkable)
│   ├── use <project> <repo>     # Set default project/repo context
│   └── clear                    # Clear defaults
│
├── branch
│   ├── list <project> <repo>            # list_branches (--filter)
│   ├── create <project> <repo> <name> --from <ref>  # create_branch
│   ├── default <project> <repo>         # get_default_branch
│   └── delete <project> <repo> <name>   # delete_branch [dangerous]
│
├── tag
│   ├── list <project> <repo>            # list_tags (--filter)
│   └── delete <project> <repo> <name>   # delete_tag [dangerous]
│
├── pr
│   ├── list <project> <repo>    # --state, --direction, --author, --draft, --order
│   ├── get <project> <repo> <id>
│   ├── create <project> <repo>  # --title, --source, --target, --description, --reviewer, --draft
│   ├── update <project> <repo> <id>  # --title, --description, --target, --reviewer, --draft
│   ├── merge <project> <repo> <id>   # --strategy
│   ├── decline <project> <repo> <id>
│   ├── reopen <project> <repo> <id>
│   ├── approve <project> <repo> <id>
│   ├── unapprove <project> <repo> <id>
│   ├── request-changes <project> <repo> <id>
│   ├── remove-request <project> <repo> <id>
│   ├── can-merge <project> <repo> <id>
│   ├── diff <project> <repo> <id>       # --context
│   ├── diffstat <project> <repo> <id>
│   ├── commits <project> <repo> <id>
│   ├── activities <project> <repo> <id>
│   ├── participants <project> <repo> <id>
│   ├── watch <project> <repo> <id>
│   ├── unwatch <project> <repo> <id>
│   ├── draft <project> <repo>           # Alias for create --draft
│   ├── publish <project> <repo> <id>
│   ├── convert-to-draft <project> <repo> <id>
│   ├── suggest-message <project> <repo> <id>
│   ├── delete <project> <repo> <id>     # [dangerous]
│   ├── comment
│   │   ├── list <project> <repo> <pr-id>
│   │   ├── get <project> <repo> <pr-id> <comment-id>
│   │   ├── add <project> <repo> <pr-id>  # --text, --file, --line, --reply-to, --blocker
│   │   ├── update <project> <repo> <pr-id> <comment-id> --text
│   │   ├── resolve <project> <repo> <pr-id> <comment-id>
│   │   ├── reopen <project> <repo> <pr-id> <comment-id>
│   │   └── delete <project> <repo> <pr-id> <comment-id>  # [dangerous]
│   └── task
│       ├── list <project> <repo> <pr-id>
│       ├── create <project> <repo> <pr-id>  # --text, --comment-id
│       ├── get <project> <repo> <pr-id> <task-id>
│       ├── update <project> <repo> <pr-id> <task-id>  # --text, --state
│       └── delete <project> <repo> <pr-id> <task-id>  # [dangerous]
│
├── commit
│   ├── list <project> <repo>    # --until, --since, --path
│   ├── get <project> <repo> <id>
│   ├── diff <project> <repo> <id>   # --context, --src-path
│   └── changes <project> <repo> <id>
│
├── file
│   ├── browse <project> <repo>       # --path, --at
│   ├── cat <project> <repo> <path>   # --at
│   ├── list <project> <repo>         # --path, --at
│   └── find <project> <repo> <pattern>
│
├── search
│   └── code <query>   # --project, --repo, --extension, --language
│
├── user
│   └── find <query>
│
├── dashboard
│   ├── list           # --state, --role, --order
│   └── inbox
│
├── attachment
│   ├── get <project> <repo> <id>
│   ├── meta <project> <repo> <id>
│   ├── save-meta <project> <repo> <id>
│   ├── delete <project> <repo> <id>       # [dangerous]
│   └── delete-meta <project> <repo> <id>  # [dangerous]
│
├── completion [bash|zsh|fish]
└── --version
```

### Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--profile` | `default` | Config profile |
| `--json` | `false` | JSON output |
| `--format` | `""` | Go template |
| `--limit` | `25` | Pagination limit |
| `--no-color` | `false` | Disable colors |

### Pagination Flags (on list commands)

| Flag | Default | Description |
|------|---------|-------------|
| `--start` | `0` | Pagination offset |
| `--limit` | `25` | Items per page |
| `--all` | `false` | Auto-paginate all results |

---

## 3. HTTP Client & API Layer

### Client

```go
type Client struct {
    baseURL    string
    token      string
    httpClient *http.Client
}
```

All requests go to `{baseURL}/rest/api/1.0/...` with `Authorization: Bearer {token}`. Special endpoints:
- `/rest/search/latest/` — code search
- `/rest/branch-utils/1.0/` — branch deletion
- `/rest/git/1.0/` — tag deletion

### Pagination

```go
type PagedResponse[T any] struct {
    Values        []T  `json:"values"`
    Size          int  `json:"size"`
    Start         int  `json:"start"`
    Limit         int  `json:"limit"`
    IsLastPage    bool `json:"isLastPage"`
    NextPageStart int  `json:"nextPageStart"`
}
```

Generic `GetPaged[T]()` for single page, `GetAll[T]()` for auto-pagination (`--all` flag).

### Error Handling

```go
type APIError struct {
    StatusCode int
    Errors     []struct {
        Context       string `json:"context"`
        Message       string `json:"message"`
        ExceptionName string `json:"exceptionName"`
    } `json:"errors"`
}
```

Exit codes: `0` success, `1` general error, `2` auth error, `3` not found, `4` validation error.

### Validation

Ported from Python `validation.py`:

| Validator | Pattern/Rule |
|-----------|-------------|
| Project key | `~?[A-Za-z0-9_]{1,128}` |
| Repo slug | `[A-Za-z0-9][A-Za-z0-9._-]*` |
| Commit ID | `[0-9a-fA-F]{4,40}` |
| Branch/tag name | No `//`, no trailing `/`, max 256 chars |
| Path | No `..`, no leading `/`, no null bytes |
| Limit | Clamped 1–1000 (default 25) |
| Context lines | Clamped 0–100 (default 10) |

---

## 4. Configuration & Authentication

### File Locations

```
~/.config/bb/config.yaml       # Profiles, defaults
~/.config/bb/credentials.yaml  # Tokens (mode 0600)
```

### Config Format

```yaml
# config.yaml
current-profile: default
profiles:
  default:
    url: https://bitbucket.example.com
    default-project: MYPROJ
    default-repo: my-repo
  staging:
    url: https://bitbucket-staging.example.com
```

```yaml
# credentials.yaml (0600)
profiles:
  default:
    token: "BBAT-xxxxxxxxxxxxx"
  staging:
    token: "BBAT-yyyyyyyyyyyyy"
```

### Resolution Order (highest wins)

1. CLI flags (`--profile`)
2. Environment variables (`BITBUCKET_URL`, `BITBUCKET_TOKEN`)
3. Config files

### Auth Commands

- `bb auth login` — Interactive: prompt URL, token, profile name. Validates by calling `/rest/api/1.0/users` endpoint.
- `bb auth login --url <url> --token <token>` — Non-interactive for automation.
- `bb auth logout [--profile <name>]` — Remove stored credentials.
- `bb auth status` — Show profile, URL, authenticated user.

### Default Context

- `bb repo use <project> <repo>` — Set defaults in active profile.
- `bb repo clear` — Clear defaults.
- When defaults are set, `<project>` and `<repo>` args become optional (explicit args override).

---

## 5. Output Formatting

### Three Modes

**Table (default):** Human-readable, colored via Lipgloss.

```
ID    TITLE                        AUTHOR       STATE   REVIEWERS        UPDATED
42    Add login endpoint           jsmith       OPEN    mjones ✓, klee   2h ago
```

Reviewer indicators: `✓` approved, `✗` needs work, `⏳` unapproved.

**JSON (`--json`):** Full API response, pretty-printed. No color, pipes cleanly to `jq`.

**Go template (`--format`):** Custom fields via Go `text/template`.

```
bb pr list PROJ repo --format '{{.ID}}\t{{.Title}}\t{{.Author.User.DisplayName}}'
```

### Behavioral Details

- **TTY detection:** When stdout is piped, auto-disable colors, use tab-separated plain text.
- **Stderr for status:** Progress messages go to stderr.
- **Exit codes:** Structured for scripting (see Section 3).

### Formatter Interface

```go
type Formatter interface {
    Format(data any) (string, error)
}
```

Factory selects `TableFormatter`, `JSONFormatter`, or `TemplateFormatter` based on flags.

---

## 6. Dangerous Operations & Safety

### Two-Tier Model

**Tier 1 — Dangerous** (branch/tag/PR/comment/task/attachment deletes):
- Interactive confirmation: type the resource name.
- `--confirm` flag skips prompt for scripting.

**Tier 2 — Destructive** (project delete, repo delete):
- Interactive confirmation: type `project/repo`.
- Scripting requires both `--confirm` and `--i-understand-this-is-destructive`.

No environment variable gates (unlike MCP server) — interactive confirmation is the appropriate safety mechanism for a human-facing CLI.

---

## 7. Key Data Types

```go
type Project struct {
    Key, Name, Description string
    Public                 bool
    Type                   string
}

type Repository struct {
    Slug, Name, Description, State, ScmID string
    Project                               Project
    Forkable                              bool
    Links                                 Links
}

type PullRequest struct {
    ID, Version                int
    Title, Description, State  string
    Draft                      bool
    Author                     Participant
    Reviewers                  []Participant
    FromRef, ToRef             Ref
    CreatedDate, UpdatedDate   int64
}

type Participant struct {
    User     User
    Role     string  // AUTHOR, REVIEWER, PARTICIPANT
    Approved bool
    Status   string  // APPROVED, UNAPPROVED, NEEDS_WORK
}

type User struct {
    Name, DisplayName, EmailAddress, Slug string
}

type Ref struct {
    ID, DisplayID, LatestCommit string
    Repository                  Repository
}

type Commit struct {
    ID, DisplayID, Message string
    Author                 Person
    AuthorTimestamp        int64
    Parents                []Commit
}

type Comment struct {
    ID, Version          int
    Text, Severity, State string
    Author               User
    Anchor               *Anchor
    Comments             []Comment
    CreatedDate          int64
}

type Anchor struct {
    Path, LineType, FileType string
    Line                     int
}

type Task struct {
    ID         int
    Text, State string
}
```

All types carry JSON struct tags mapping to Bitbucket API field names. `Version` fields support optimistic locking on updates/deletes.

---

## 8. Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/spf13/viper` | Config management |
| `github.com/charmbracelet/lipgloss` | Terminal styling |
| `github.com/charmbracelet/x/term` | TTY detection |
| `golang.org/x/term` | Password/token input |

HTTP via Go stdlib `net/http`. No external HTTP client.

### Build

```makefile
VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -s -w -X main.version=$(VERSION)

build:   go build -ldflags "$(LDFLAGS)" -o bin/bb ./cmd/bb
install: go install -ldflags "$(LDFLAGS)" ./cmd/bb
test:    go test ./... -race -cover
lint:    golangci-lint run
```

Single static binary. Cross-compile via `GOOS`/`GOARCH`.

---

## 9. Testing Strategy

- **Client tests:** `net/http/httptest` server to verify request construction and response parsing.
- **Validation tests:** Table-driven tests for every validator with valid/invalid inputs.
- **Command tests:** Execute Cobra commands against mock client, assert stdout output.
- **Output tests:** Verify each formatter produces expected strings.
- **Integration tests:** `//go:build integration` tag, requires live Bitbucket Server + credentials.

Not tested: Lipgloss visual styling, Viper/Cobra internals (well-tested upstream).
