# C4 — C2 Control Center

[![Go version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](https://go.dev/dl/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

C4 is a command-line control center for deploying, managing, and destroying Command & Control (C2) frameworks from one interface. It targets [Mythic](https://github.com/its-a-feature/Mythic) and Sliver, orchestrating containers via Docker Compose and talking to a GraphQL (Hasura) backend for configuration and state.

Built for red teams and adversary-emulation work, C4 turns the repetitive lifecycle of standing up and tearing down C2 infrastructure — deploy, start, stop, destroy — into a few deterministic commands, with listener, callback, and payload management on top.

## Features

- **Multi-framework control plane** — a provider abstraction (`pkg/provider`) with a Mythic provider and a generic exec provider; new backends plug in without touching the CLI.
- **Full lifecycle management** — `deploy`, `start`, `stop`, and `destroy` for C2 instances.
- **Operator tooling** — manage listeners, callbacks, and payloads per instance.
- **Auto-detection** — `c4 detect` scans the host for already-installed C2 frameworks.
- **Compose-native orchestration** — instances are rendered from Go templates into Docker Compose projects (`resources/compose`) and driven through the `docker compose` CLI.
- **Scriptable output** — TOML configuration, JSON log output, and verbose/debug logging for automation and CI.
- **Clean teardown** — `destroy` removes containers, networks, and volumes so every run starts from a known-good state.

## Install

```bash
go install github.com/CR1MS0N-Operator/c4@latest
```

Requires Go 1.26+.

## Quick Start

```bash
# Generate a default configuration (~/.c4/c4.toml)
c4 init

# Deploy a Mythic instance
c4 deploy mythic

# Show instance status
c4 status

# List active listeners
c4 listener list

# Tear everything down
c4 destroy mythic
```

## Command Reference

| Command            | Description                              |
|--------------------|------------------------------------------|
| `c4 init`          | Generate a default configuration at `~/.c4/c4.toml` |
| `c4 deploy <c2>`   | Deploy or start a C2 instance            |
| `c4 destroy <c2>`  | Destroy a C2 instance                    |
| `c4 status [c2]`   | Show C2 instance status (default: mythic)|
| `c4 start <c2>`    | Start an existing C2 instance            |
| `c4 stop <c2>`     | Stop a running C2 instance               |
| `c4 listener`      | Manage C2 listeners                      |
| `c4 callback`      | Manage C2 callbacks                      |
| `c4 payload`       | Manage C2 payloads                       |
| `c4 config`        | Manage C4 configuration (`init`, `set`, `view`) |
| `c4 detect`        | Auto-detect installed C2 frameworks      |

## Architecture

```
                    ┌──────────────────────────────────┐
                    │            c4 CLI (cmd/)          │
                    │  Cobra command tree · config ·    │
                    │  zerolog logging (console/JSON)   │
                    └───────────────┬──────────────────┘
                                    │
                        ┌───────────┴───────────┐
                        │   pkg/provider        │
                        │  Provider interface   │
                        │  (Type · Status lifecycle) │
                        └───────┬───────────┬───┘
                    ┌───────────┴──┐   ┌────┴──────────┐
                    │ pkg/mythic   │   │ pkg/execprovider│
                    │ Mythic API   │   │ generic exec   │
                    └───────┬──────┘   └───────────────┘
                            │
          ┌─────────────────┼──────────────────┐
          │                 │                  │
  ┌───────┴───────┐  ┌──────┴───────┐  ┌──────┴───────┐
  │  pkg/docker   │  │ pkg/graphql │  │  pkg/detect  │
  │ compose CLI   │  │ Hasura      │  │ framework    │
  │ orchestration │  │ client      │  │ detection    │
  └───────┬───────┘  └─────────────┘  └──────────────┘
          │
  resources/compose/mythic.tmpl.yml   (rendered by pkg/template)
```

| Package           | Responsibility                                        |
|-------------------|-------------------------------------------------------|
| `cmd/`            | Cobra command tree, persistent flags, config loading  |
| `pkg/provider`    | Provider interface and lifecycle types shared by backends |
| `pkg/mythic`      | Mythic C2 backend (deploy, status, listener, callback, payload) |
| `pkg/execprovider`| Generic backend that runs configured start/stop commands |
| `pkg/docker`      | Shells out to the `docker compose` CLI for lifecycle ops |
| `pkg/graphql`     | GraphQL client for the Hasura backend                 |
| `pkg/config`      | TOML config with safe defaults (`~/.c4/c4.toml`)      |
| `pkg/detect`      | Auto-detection of installed C2 frameworks             |
| `pkg/template`    | Renders Compose templates from `resources/compose`    |

## Requirements

- Go 1.26+
- Docker Engine 24+ with Docker Compose v2
- A Hasura GraphQL endpoint (configured via `c4 init`)

## Ecosystem

C4 is part of the CR1MS0N-Operator toolset:

- [Veil](https://github.com/CR1MS0N-Operator/veil) — production-grade red team infrastructure as code (multi-node WireGuard mesh, Mythic C2)
- [NightForge](https://github.com/CR1MS0N-Operator/nightforge) — reproducible Arch Linux red team operator workstation
- [Lantern](https://github.com/CR1MS0N-Operator/ACLGuard-Active-Directory-Permission-Auditor) — Active Directory identity exposure validation

## License

MIT — see [LICENSE](LICENSE).
