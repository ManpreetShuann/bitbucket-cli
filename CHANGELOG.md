# Changelog

All notable changes to this project are documented in [GitHub Releases](https://github.com/ManpreetShuann/bitbucket-cli/releases).

Release notes are auto-generated from [Conventional Commits](https://www.conventionalcommits.org/) via [GoReleaser](https://goreleaser.com/).

## [v0.1.2](https://github.com/ManpreetShuann/bitbucket-cli/releases/tag/v0.1.2) — 2026-03-24

### Fixes

- Handle all unchecked error returns flagged by errcheck linter
- Build golangci-lint from source to match current Go version

## [v0.1.1](https://github.com/ManpreetShuann/bitbucket-cli/releases/tag/v0.1.1) — 2026-03-24

### Fixes

- Lower go directive to 1.22 for golangci-lint compatibility
- Auth status/login now show the authenticated user, not first alphabetical user

## [v0.1.0](https://github.com/ManpreetShuann/bitbucket-cli/releases/tag/v0.1.0) — 2026-03-24

### Initial Release

- 66+ commands across 12 resource groups (projects, repos, branches, tags, PRs, comments, tasks, commits, files, search, users, dashboard, attachments)
- Profile-based authentication with multi-instance support
- Flexible output: table, JSON, Go templates
- Two-tier safety model for dangerous and destructive operations
- Shell completions for Bash, Zsh, Fish, PowerShell
- Automatic retry with exponential backoff on 429/503
- Debug mode for HTTP request/response inspection
