# c4 — Agent Rules

## Purpose
C4 is the **C2 Control Center** for the CR1MS0N continuous adversarial validation platform — a CLI to deploy, manage, and destroy Command & Control frameworks (Mythic, Sliver) via Docker Compose + Hasura GraphQL backend.

Pairs with [[nightforge]] (measurement/mobilization) and [[veil]] (validation substrate).

## Stack
- **Language:** Go 1.26+
- **CLI:** Cobra
- **Orchestration:** Docker Compose (rendered from Go templates in `resources/compose/`)
- **Backend:** Hasura GraphQL
- **Providers:** Mythic, exec (generic)
- **Config:** TOML

## Agent Architecture
- **Brain:** Hermes Desktop (planning, review)
- **Executor:** dsh web API (deepseek) — code generation, refactoring
- **Local:** Pi Agent (Qwen3.8-27B) — for privacy-sensitive code review

## Commands
```bash
go build ./cmd/c4/
go vet ./...
go test ./...
./c4 detect              # scan host for installed C2 frameworks
./c4 deploy <instance>   # render + docker compose up
./c4 start <instance>
./c4 stop <instance>
./c4 destroy <instance>  # full teardown
```

## Repository Layout
```
cmd/            CLI entry points
pkg/
  provider/     provider abstraction (Mythic, exec)
  lifecycle/    deploy/start/stop/destroy
  compose/      Docker Compose rendering
  graphql/      Hasura client
resources/
  compose/      Go templates for Docker Compose projects
main.go
go.mod
```

## Conventions
- **Commits:** Conventional Commits, one logical change per commit
- **Go style:** `gofmt`, `go vet ./...`, no new globals without justification
- **Providers:** New C2 backends plug in via `pkg/provider/` — don't touch CLI
- **Compose:** All C2 instances rendered from Go templates, never hand-edited
- **Tests:** Run `go test ./...` before every commit

## Guardrails
- **Do not push** without operator approval
- **Do not destroy** live C2 instances without explicit operator confirmation
- **Do not modify** `resources/compose/` templates without updating provider tests
- **Do not commit** secrets, API tokens, or live C2 config to repo
- **Documentation-only tasks** stay documentation-only

## Subagent-Driven Development
For multi-step work (provider addition, lifecycle changes, compose refactor):
1. **Decompose:** Plan into independent steps (read provider spec, draft code, write tests, update docs)
2. **Spawn parallel leaves:** Dispatch independent steps as parallel dsh sessions (one per step)
3. **Synthesize:** Parent (Hermes) collects `session.history` from each, verifies, integrates
4. **Verify:** `go build ./... && go test ./...` must pass before commit

## Escalation Triggers
- Provider API change (Mythic/Sliver breaking changes)
- Hasura schema change
- Docker Compose version upgrade
- New C2 framework integration request
- Security finding in C2 deployment code

## Reference
- Platform context: [[ai-lab-vault|AGENTS.md]]
- NightForge: [[nightforge]]
- Veil: [[veil]]
