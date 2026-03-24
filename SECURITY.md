# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in `bb`, **please do not open a public issue**.

Instead, use [GitHub's private vulnerability reporting](https://github.com/ManpreetShuann/bitbucket-cli/security/advisories/new) to submit your report. This ensures the issue stays confidential until a fix is available.

### What to include

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Response timeline

- **Acknowledgement**: within 48 hours
- **Initial assessment**: within 1 week
- **Fix and disclosure**: coordinated with the reporter

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | Yes       |
| < latest | No       |

We only patch the most recent release. Users are encouraged to stay up to date.

## Scope

This policy covers the `bb` CLI binary and its source code. It does **not** cover:

- Bitbucket Server itself
- Third-party dependencies (report those upstream)

## Credential Handling

`bb` stores personal access tokens in `~/.config/bb/credentials.yaml` with `0600` file permissions. Tokens are never logged, printed, or transmitted to any endpoint other than the configured Bitbucket Server instance. If you find a scenario where credentials are exposed, please report it immediately.
