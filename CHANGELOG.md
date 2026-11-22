# Changelog

All notable changes to C4. Format follows [Keep a Changelog](https://keepachangelog.com/); versioning is semantic. Commit hashes reference `main`.

## [Unreleased]

### Added
- CONTRIBUTING.md, CHANGELOG.md, and root SECURITY.md for cross-repo consistency
- Deployment warning in `resources/compose/mythic.tmpl.yml` — change default passwords before deployment
- ARCHITECTURE.md — package layout, provider abstraction, and data flow; linked from README
- WORKTREES.md — dispatch isolation runbook for treehouse worktrees (branch-before-commit, commit-in-worktree, reachability check)
- `treehouse.toml` — explicit worktree pool config (`root`, `max_trees`, `base_branch = "main"`)

### Changed
- Go and License badges switched to `flat-square` style

### Security
- Mythic connections now verify the server's TLS certificate instead of accepting any certificate. Mythic ships a self-signed certificate, so set `cert_fingerprint` or `ca_file` under `[mythic]` (or `insecure_skip_verify = true` to opt out) — see README "Mythic TLS"
- Removed the silent plain-HTTP retry on `server_port`. The retry is now opt-in (`server_port_fallback = true`), uses the same scheme and TLS settings as `api_port`, and prints a warning
- The GraphQL client no longer follows redirects to another origin, so `x-hasura-admin-secret` is never forwarded
- `c4.toml` is always written with mode `0600` (`c4 config set` and `config.Save` used `0644`), an existing file is tightened on rewrite, and C4 creates `~/.c4` and `~/.c4/providers` with mode `0700`
- `c4 config set` no longer takes a value on the command line for keys containing `secret`: omit the value to be prompted without echo, or pass `-` to read it from stdin. The value is not printed back

### Fixed
- `c4 config set` keeps the other keys in `c4.toml` instead of rewriting the file with only the key being set
- `c4 detect` now returns an empty result set instead of nil — `c4 detect --json` emits `[]` rather than `null` when nothing is found

## [0.1.0] — 2026-07-26

First tagged release. C2 Control Center for the CR1MS0N platform — deploy, manage, and destroy Mythic C2 instances from one CLI.

### Added
- Listener management — List, Start, Stop via Mythic API (`9584666`)
- Callback and payload listing via Mythic API (`87eeb6e`)
- Exec provider for custom C2s + `c4 detect` host auto-detection (`9af4aea`, `3517734`)
- Compose-native orchestration: instances rendered from Go templates in `resources/compose` and driven through the `docker compose` CLI
- TOML configuration with safe defaults at `~/.c4/c4.toml` (`c4 init`)
- JSON log output and verbose/debug logging for automation and CI
- README with install and usage (`fd0bef5`)
- Hasura secret override path documented (`e91ceec`)

### Changed
- Migrated module path to `github.com/CR1MS0N-Operator/c4` (`e6d5b8d`)
- Bumped version from dev to v0.1.0 (`7237830`)

### Fixed
- Removed tracked `c4` binary, added to `.gitignore` (`d6745ed`)

### Security
- Added MIT LICENSE and `.github/SECURITY.md` security policy (`6178773`)

## 2026-08-11 → 2026-08-22 — Documentation & CI

- README overhaul with architecture diagram, command reference, and ecosystem links
- `.gitignore` updated for build artifacts and `graphify-out/`
- Final pre-wipe documentation sync (S216)

---

*Individual commits are atomic by design — use `git log --format='%h %ad %s' --date=short` for the full breakdown.*
