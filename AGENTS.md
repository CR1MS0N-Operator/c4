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

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:46cd31e7 -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/core-concepts/sync-concepts.md for details and anti-patterns.

## Agent Context Profiles

The managed Beads block is task-tracking guidance, not permission to override repository, user, or orchestrator instructions.

- **Conservative (default)**: Use `bd` for task tracking. Do not run git commits, git pushes, or Dolt remote sync unless explicitly asked. At handoff, report changed files, validation, and suggested next commands.
- **Minimal**: Keep tool instruction files as pointers to `bd prime`; use the same conservative git policy unless active instructions say otherwise.
- **Team-maintainer**: Only when the repository explicitly opts in, agents may close beads, run quality gates, commit, and push as part of session close. A current "do not commit" or "do not push" instruction still wins.

## Session Completion

This protocol applies when ending a Beads implementation workflow. It is subordinate to explicit user, repository, and orchestrator instructions.

1. **File issues for remaining work** - Create beads for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **Handle git/sync by active profile**:
   ```bash
   # Conservative/minimal/default: report status and proposed commands; wait for approval.
   git status

   # Team-maintainer opt-in only, unless current instructions forbid it:
   git pull --rebase
   bd dolt push
   git push
   git status
   ```
5. **Hand off** - Summarize changes, validation, issue status, and any blocked sync/commit/push step

**Critical rules:**
- Explicit user or orchestrator instructions override this Beads block.
- Do not commit or push without clear authority from the active profile or the current user request.
- If a required sync or push is blocked, stop and report the exact command and error.
<!-- END BEADS INTEGRATION -->

<!-- BEGIN BEADS CODEX SETUP: generated by bd setup codex -->
## Beads Issue Tracker

Use Beads (`bd`) for durable task tracking in repositories that include it. Use the `beads` skill at `.agents/skills/beads/SKILL.md` (project install) or `~/.agents/skills/beads/SKILL.md` (global install) for Beads workflow guidance, then use the `bd` CLI for issue operations.

### Quick Reference

```bash
bd ready                # Find available work
bd show <id>            # View issue details
bd update <id> --claim  # Claim work
bd close <id>           # Complete work
bd prime                # Refresh Beads context
```

### Rules

- Use `bd` for all task tracking; do not create markdown TODO lists.
- Run `bd prime` when Beads context is missing or stale. Codex 0.129.0+ can load Beads context automatically through native hooks; use `/hooks` to inspect or toggle them.
- Keep persistent project memory in Beads via `bd remember`; do not create ad hoc memory files.

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/core-concepts/sync-concepts.md for details and anti-patterns.
<!-- END BEADS CODEX SETUP -->
