# Contributing to C4

Thanks for contributing. C4 is the **C2 Control Center** of the CR1MS0N continuous adversarial validation platform — a Go CLI that deploys, manages, and tears down Mythic and other C2 frameworks via Docker Compose and a Hasura GraphQL backend.

## Prerequisites

- **Go 1.26+** (go.mod targets 1.26)
- Docker Engine 24+ with Docker Compose v2
- A Hasura GraphQL endpoint for full lifecycle testing (optional for build/vet)

## Development setup

```bash
# Clone
git clone https://github.com/CR1MS0N-Operator/c4.git
cd c4

# Build the CLI
go build ./cmd/c4/

# Vet + test
go vet ./...
go test ./...
```

The module path is `github.com/CR1MS0N-Operator/c4`.

## Building and testing

```bash
# Build
go build ./...

# Vet
go vet ./...

# Test
go test ./...

# Smoke run (requires Docker)
./c4 init
./c4 detect
./c4 deploy mythic
./c4 status
./c4 destroy mythic
```

Run `go vet ./... && go test ./...` before every commit.

## Code style

- `gofmt`/`gofumpt` clean, `go vet` clean — no lint suppressions without a comment.
- Errors are values: wrap with `fmt.Errorf("...: %w", err)`; never swallow errors silently.
- New C2 backends plug in via `pkg/provider` — don't touch the CLI to add a provider.
- All C2 instances are rendered from Go templates in `resources/compose`; never hand-edit generated Compose files.
- Config lives in `pkg/config`; defaults are written to `~/.c4/c4.toml` by `c4 init`.

## Commit conventions

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(provider): add Sliver provider
fix(template): correct Mythic port binding
docs: add listener command examples
chore: bump Go 1.26
```

- **Micro-commits:** one logical change per commit. No mixed-purpose commits.
- Subject ≤ 72 chars, imperative mood, no trailing period.
- Body explains *why* when the subject isn't enough.
- Do not push without explicit operator approval.

## Branches and PRs

1. Create a feature branch from `main` (`git checkout -b feat/<name>`).
2. Make micro-commits.
3. Verify: `go build ./... && go vet ./... && go test ./...`.
4. Review the diff before committing/PRing.
5. Open a PR against `main` with a summary of what changed and verification evidence.

## Guardrails

- **Do not destroy** live C2 instances without explicit operator confirmation.
- **Do not modify** `resources/compose/` templates without updating provider tests.
- **Do not commit** secrets, API tokens, or live C2 configuration.
- **Do not commit** default passwords — the Compose templates contain placeholders; override them via environment variables or the C4 config at deploy time.
- **Documentation-only changes** stay documentation-only.

## Reporting issues

- Bugs: include the command run, environment, and the exact output.
- Security findings: do **not** file a public issue — follow [SECURITY.md](SECURITY.md).
