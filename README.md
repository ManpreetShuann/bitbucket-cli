# bb — Bitbucket Server CLI

> A fast, ergonomic command-line interface for **Bitbucket Server** (formerly Stash).

`Go` | `66+ commands` | `12 resource groups` | `MIT License`

---

Wraps the Bitbucket Server REST API with Git-style subcommands, profile-based authentication, flexible output formatting, and a two-tier safety model for dangerous operations. Built for engineers who live in the terminal.

## Highlights

| | |
|---|---|
| 🔌 **Full API Coverage** | Projects, repos, branches, tags, PRs (with comments, tasks, activities), commits, files, code search, dashboard, attachments |
| 🔑 **Profile-Based Auth** | Manage multiple Bitbucket Server instances with named profiles |
| 📊 **Flexible Output** | Table (default), JSON (`--json`), or Go templates (`--format`) |
| 🛡️ **Two-Tier Safety** | `--confirm` for dangerous ops, `--confirm --i-understand-this-is-destructive` for irreversible ones |
| 🐚 **Shell Completion** | Bash, Zsh, Fish, and PowerShell |
| 🔁 **Automatic Retry** | Exponential backoff on 429/503 responses |
| 🐛 **Debug Mode** | `--debug` to inspect raw HTTP request/response on stderr |

## Installation

### From Source

Requires **Go 1.22+**:

```bash
git clone https://github.com/manu/bb.git
cd bb
make build        # builds binary to bin/bb
make install      # installs to $GOPATH/bin
```

The version is automatically set from `git describe --tags` at build time.

## Quick Start

### 1. Authenticate

```bash
# Interactive — prompts for URL, token, default project & repo
bb auth login

# Non-interactive
bb auth login --url https://bitbucket.example.com --token BBAT-xxxxxxxxxxxx

# Check current auth status
bb auth status
```

### 2. Set a Default Project & Repo Context

```bash
# Set defaults so you don't have to pass --project / --repo every time
bb repo use MYPROJ my-repo

# Clear the saved context
bb repo clear
```

### 3. Start Working

```bash
# List open pull requests
bb pr list

# Create a pull request
bb pr create --from feature/my-branch --to main --title "Add caching layer"

# Approve and merge
bb pr approve 42
bb pr merge 42 --confirm

# View your personal dashboard
bb dashboard list

# Search across all repos
bb search code "func main" --lang go
```

### 4. Use JSON Output with jq

```bash
# Get PR IDs and titles as JSON, then filter with jq
bb pr list --json | jq '.[].title'

# Extract reviewers from a specific PR
bb pr get 42 --json | jq '.reviewers[].user.displayName'

# Count open PRs
bb pr list --json | jq length
```

## Commands

`bb` provides **66+ commands** across **12 resource groups** plus shell completions:

### Authentication

```
bb auth login              Log in to a Bitbucket Server instance
bb auth logout             Remove saved credentials
bb auth status             Show current authentication state
```

### Projects

```
bb project list            List all projects
bb project get             Get project details
bb project delete          Delete a project (Tier 2)
```

### Repositories

```
bb repo list               List repositories in a project
bb repo get                Get repository details
bb repo create             Create a new repository
bb repo delete             Delete a repository (Tier 2)
bb repo use                Set default project and repo context
bb repo clear              Clear saved project/repo context
```

### Branches

```
bb branch list             List branches
bb branch create           Create a new branch
bb branch default          Get or set the default branch
bb branch delete           Delete a branch (Tier 1)
```

### Tags

```
bb tag list                List tags
bb tag delete              Delete a tag (Tier 1)
```

### Pull Requests

```
bb pr list                 List pull requests
bb pr get                  Get pull request details
bb pr create               Create a pull request
bb pr update               Update PR title, description, or reviewers
bb pr merge                Merge a pull request (Tier 1)
bb pr decline              Decline a pull request (Tier 1)
bb pr reopen               Reopen a declined pull request (Tier 1)
bb pr delete               Delete a pull request (Tier 2)
bb pr approve              Approve a pull request
bb pr unapprove            Remove your approval
bb pr request-changes      Request changes on a pull request
bb pr remove-request       Remove your request for changes
bb pr can-merge            Check merge prerequisites
bb pr diff                 View the pull request diff
bb pr diff-stat            View diff statistics
bb pr commits              List commits in a pull request
bb pr activities           List PR activity feed
bb pr participants         List PR participants and roles
bb pr watch                Watch a pull request for notifications
bb pr unwatch              Stop watching a pull request
bb pr draft                Create a pull request as draft
bb pr convert-to-draft     Convert an existing PR to draft
bb pr publish              Publish a draft pull request
bb pr suggest-message      Get an AI-suggested merge commit message
```

### PR Comments

```
bb pr comment list         List comments on a pull request
bb pr comment get          Get a specific comment
bb pr comment add          Add a comment to a pull request
bb pr comment update       Edit an existing comment
bb pr comment resolve      Resolve a comment thread
bb pr comment reopen       Reopen a resolved comment
bb pr comment delete       Delete a comment (Tier 1)
```

### PR Tasks

```
bb pr task list            List tasks on a pull request
bb pr task create          Create a task on a pull request
bb pr task get             Get task details
bb pr task update          Update a task (text or state)
bb pr task delete          Delete a task (Tier 1)
```

### Commits

```
bb commit list             List commits in a repository
bb commit get              Get commit details
bb commit diff             View a commit's diff
bb commit changes          List files changed in a commit
```

### Files

```
bb file browse             Browse the file tree at a path/revision
bb file cat                Print file contents to stdout
bb file list               List files at a path
bb file find               Find files by name pattern
```

### Code Search

```
bb search code             Search code across repositories
```

### Users

```
bb user find               Find users by name or email
```

### Dashboard

```
bb dashboard list          List pull requests on your dashboard
bb dashboard inbox         List your review inbox
```

### Attachments

```
bb attachment get          Download an attachment
bb attachment meta         Get attachment metadata
bb attachment save-meta    Save metadata for an attachment
bb attachment delete       Delete an attachment (Tier 1)
bb attachment delete-meta  Delete attachment metadata (Tier 1)
```

### Shell Completions

```
bb completion bash         Generate Bash completion script
bb completion zsh          Generate Zsh completion script
bb completion fish         Generate Fish completion script
bb completion powershell   Generate PowerShell completion script
```

## Global Flags

| Flag | Description |
|---|---|
| `--profile <name>` | Config profile to use (default: `default`) |
| `--json` | Output as JSON |
| `--format <template>` | Go template for custom output |
| `--no-color` | Disable colored output |
| `--debug` | Print HTTP request/response details to stderr |

## Pagination

All list commands support pagination:

```bash
bb pr list --limit 50              # first 50 results
bb pr list --limit 25 --page 2     # page 2 of 25-per-page
bb pr list --all                   # fetch every page automatically
```

Default limit is **25 results** per page.

## Output Formatting

### Table (Default)

```bash
bb pr list
```
```
┌────┬───────────────────┬────────┬──────────┐
│ ID │ TITLE             │ STATE  │ AUTHOR   │
├────┼───────────────────┼────────┼──────────┤
│ 42 │ Add caching layer │ OPEN   │ jdoe     │
│ 38 │ Fix timeout bug   │ MERGED │ asmith   │
└────┴───────────────────┴────────┴──────────┘
```

### JSON

```bash
bb pr list --json
```
```json
[
  {"id": 42, "title": "Add caching layer", "state": "OPEN", "author": "jdoe"},
  {"id": 38, "title": "Fix timeout bug", "state": "MERGED", "author": "asmith"}
]
```

### Go Template

```bash
bb pr list --format '{{.ID}} {{.Title}} ({{.State}})'
```
```
42 Add caching layer (OPEN)
38 Fix timeout bug (MERGED)
```

## Configuration

### File Locations

```
~/.config/bb/
├── config.yaml        # Profiles, defaults, and preferences
└── credentials.yaml   # Tokens (file permissions: 0600)
```

### Config File Format

```yaml
default_profile: default
profiles:
  default:
    url: https://bitbucket.example.com
    default_project: MYPROJ
    default_repo: my-repo
  staging:
    url: https://bitbucket-staging.example.com
    default_project: STAGE
```

### Environment Variables

| Variable | Description |
|---|---|
| `BITBUCKET_URL` | Server URL |
| `BITBUCKET_TOKEN` | Personal access token |

### Resolution Order

Settings are resolved with the following priority (highest first):

1. **CLI flags** (`--profile`, `--json`, etc.)
2. **Environment variables** (`BITBUCKET_URL`, `BITBUCKET_TOKEN`)
3. **Config files** (`~/.config/bb/config.yaml`, `credentials.yaml`)

## Safety Model

Operations that modify or destroy data require explicit confirmation to prevent accidents.

### Tier 1 — Dangerous

Reversible or low-blast-radius operations. Requires `--confirm`:

| Operation | Example |
|---|---|
| Merge PR | `bb pr merge 42 --confirm` |
| Decline/reopen PR | `bb pr decline 42 --confirm` |
| Delete branch | `bb branch delete feature/old --confirm` |
| Delete tag | `bb tag delete v0.1.0 --confirm` |
| Delete PR comment | `bb pr comment delete 42 100 --confirm` |
| Delete PR task | `bb pr task delete 42 200 --confirm` |
| Delete attachment | `bb attachment delete 42 --confirm` |
| Delete attachment meta | `bb attachment delete-meta 42 --confirm` |

### Tier 2 — Destructive

Irreversible operations with high blast radius. Requires **both** `--confirm` and `--i-understand-this-is-destructive`:

| Operation | Example |
|---|---|
| Delete project | `bb project delete MYPROJ --confirm --i-understand-this-is-destructive` |
| Delete repository | `bb repo delete MYPROJ my-repo --confirm --i-understand-this-is-destructive` |
| Delete pull request | `bb pr delete 42 --confirm --i-understand-this-is-destructive` |

## Shell Completions

### Bash

```bash
# Add to ~/.bashrc
eval "$(bb completion bash)"
```

### Zsh

```bash
# Add to ~/.zshrc
eval "$(bb completion zsh)"

# Or generate a file
bb completion zsh > "${fpath[1]}/_bb"
```

### Fish

```bash
bb completion fish | source

# Or persist
bb completion fish > ~/.config/fish/completions/bb.fish
```

## Debug Mode

Use `--debug` to inspect the raw HTTP requests and responses sent to the Bitbucket Server API:

```bash
bb pr list --debug
```

```
>>> GET https://bitbucket.example.com/rest/api/1.0/projects/MYPROJ/repos/my-repo/pull-requests?limit=25
>>> Authorization: Bearer BBAT-****
<<< 200 OK (142ms)
<<< Content-Type: application/json
```

Debug output is written to **stderr**, so it won't interfere with JSON piping:

```bash
bb pr list --json --debug 2>debug.log | jq '.[]'
```

## Development

```bash
make build     # Build binary to bin/bb
make test      # Run tests with race detector and coverage
make lint      # Run golangci-lint
make clean     # Remove build artifacts
```

## License

[MIT](LICENSE)
