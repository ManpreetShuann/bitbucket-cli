# Contributing to bb

Thanks for your interest in contributing! Here's everything you need to get started.

## Prerequisites

- **Go 1.22+** — [install](https://go.dev/dl/)
- **golangci-lint** — `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- A running **Bitbucket Server** instance (for manual testing)

## Setup

```bash
git clone https://github.com/ManpreetShuann/bitbucket-cli.git
cd bitbucket-cli
go mod download
```

## Common tasks

| Command | What it does |
|---------|-------------|
| `make build` | Compile binary to `bin/bb` |
| `make install` | Install `bb` to `$GOPATH/bin` |
| `make test` | Run tests with race detector and coverage |
| `make lint` | Run `golangci-lint` |
| `make clean` | Remove `bin/` |

## Project structure

```
cmd/bb/          # main entrypoint
internal/
  client/        # Bitbucket Server REST API client
  cmd/           # Cobra command definitions
  config/        # Profile-based auth & config
  output/        # Table / JSON / template formatters
  validation/    # Input validation helpers
```

## Making changes

1. Fork the repo and create a branch: `git checkout -b feat/my-feature`
2. Make your changes and add tests where appropriate
3. Run `make test && make lint`
4. Open a pull request against `master`

## Commit style

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add bb tag list --format flag
fix: handle 401 on expired token gracefully
docs: update auth login example
chore: bump golangci-lint to v1.60
```

## Adding a new command

1. Add the API method to `internal/client/`
2. Add the Cobra command in `internal/cmd/` (follow existing patterns)
3. Register it in the parent command's `init()` or `NewXxxCmd()`
4. Add a test in `internal/client/` for the API call

## Reporting issues

Use the GitHub issue templates — bug reports need `--debug` output and your `bb --version`.

## License

By contributing you agree your code will be released under the [MIT License](LICENSE).
