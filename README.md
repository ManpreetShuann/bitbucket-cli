# bb — Bitbucket Server CLI

A fast, ergonomic command-line interface for Bitbucket Server (formerly Stash). Wraps 66 REST API operations with Git-style subcommands, profile-based auth, flexible output formatting, and a two-tier safety model for dangerous operations.

## Features

- **Full API Coverage** — Projects, repos, branches, tags, pull requests (with comments, tasks, activities), commits, files, code search, dashboard, attachments
- **Profile-Based Auth** — Multiple server configurations with named profiles
- **Flexible Output** — Table (default), JSON (`--json`), or Go templates (`--format`)
- **Safety Model** — Two-tier confirmation for dangerous and destructive operations
- **Shell Completion** — Bash, Zsh, Fish, and PowerShell
- **Retry Logic** — Automatic retry on 429/503 with exponential backoff

## Installation

### From Source

```bash
# Requires Go 1.22+
git clone https://github.com/manu/bb.git
cd bb
make build        # binary at bin/bb
make install      # installs to $GOPATH/bin
```

## Quick Start

### 1. Configure Authentication

```bash
# Interactive setup
bb auth login

# Or set environment variables
export BITBUCKET_URL=https://bitbucket.example.com
export BITBUCKET_TOKEN=your-personal-access-token
```

### 2. Set Default Project/Repo

```bash
bb auth login
# Follow prompts to set URL, token, default project & repo
```

### 3. Start Using

```bash
# List your pull requests
bb dashboard

# Create a pull request
bb pr create --from feature/my-branch --to main --title "My PR"

# List branches
bb branch list

# Search code
bb search "func main" --lang go
```

## Commands

```
bb auth        Manage authentication (login, logout, status, switch)
bb project     Manage projects (list, get, delete)
bb repo        Manage repositories (list, get, create, delete)
bb branch      Manage branches (list, default, create, delete)
bb tag         Manage tags (list, delete)
bb pr          Manage pull requests (25 subcommands)
  bb pr list / get / create / draft / merge / approve / decline / ...
  bb pr comment   (add, get, list, update, delete, resolve, reopen)
  bb pr task      (create, get, list, update, delete)
bb commit      Manage commits (list, get, diff, changes)
bb file        Browse and view files (browse, cat, list)
bb search      Search code (code, find-file)
bb user        User operations (find)
bb dashboard   Dashboard and inbox (list, inbox)
bb attachment  Manage attachments (get, save, meta, delete)
bb completion  Shell completions (bash, zsh, fish, powershell)
```

## Global Flags

| Flag | Description |
|------|-------------|
| `--profile` | Config profile to use (default: `default`) |
| `--json` | Output as JSON |
| `--format` | Go template for custom output |
| `--no-color` | Disable colored output |
| `--debug` | Print HTTP request/response details to stderr |

## Configuration

Config files are stored in `~/.config/bb/`:

```
~/.config/bb/
├── config.yaml        # Profiles and defaults
└── credentials.yaml   # Tokens (mode 0600)
```

### Config File Format

```yaml
default_profile: default
profiles:
  default:
    url: https://bitbucket.example.com
    default_project: MYPROJ
    default_repo: my-repo
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `BITBUCKET_URL` | Server URL (overrides config) |
| `BITBUCKET_TOKEN` | Personal access token (overrides config) |

Resolution order: CLI flags → environment variables → config file.

## Output Formatting

### Table (Default)

```bash
bb pr list
# ┌────┬───────────────┬────────┬──────────┐
# │ ID │ TITLE         │ STATE  │ AUTHOR   │
# ├────┼───────────────┼────────┼──────────┤
# │ 42 │ Add feature X │ OPEN   │ jdoe     │
# └────┴───────────────┴────────┴──────────┘
```

### JSON

```bash
bb pr list --json
# [{"id":42,"title":"Add feature X","state":"OPEN",...}]
```

### Go Template

```bash
bb pr list --format '{{.ID}} {{.Title}} ({{.State}})'
# 42 Add feature X (OPEN)
```

## Safety Model

Operations are classified into two tiers:

### Tier 1 — Dangerous (Reversible)

Requires `--confirm` flag:
- Merge, decline, reopen, delete PR comments/tasks

```bash
bb pr merge 42 --confirm
```

### Tier 2 — Destructive (Irreversible)

Requires both `--confirm` and `--i-understand-this-is-destructive`:
- Delete branches, tags, repositories, projects, PRs

```bash
bb repo delete PROJ my-repo --confirm --i-understand-this-is-destructive
```

## Pagination

List commands support pagination:

```bash
bb pr list --page 1 --limit 25
bb pr list --page 2 --limit 50
```

Default limit is 25 results per page.

## Development

```bash
make build     # Build binary
make test      # Run tests with race detector
make lint      # Run golangci-lint
make clean     # Remove build artifacts
```

## License

MIT
