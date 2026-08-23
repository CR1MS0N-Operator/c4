# C4 — C2 Control Center

[![Go version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](https://go.dev/dl/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

C4 is a command-line control center for deploying, managing, and destroying Command & Control (C2) frameworks from one interface. It targets [Mythic](https://github.com/its-a-feature/Mythic) and Sliver, orchestrating containers via Docker Compose and talking to a GraphQL (Hasura) backend for configuration and state.

Built for adversary-emulation across the [CR1MS0N continuous adversarial validation platform](https://github.com/CR1MS0N-Operator/veil), C4 turns the repetitive lifecycle of standing up and tearing down C2 infrastructure — deploy, start, stop, destroy — into a few deterministic commands, with listener, callback, and payload management on top.

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

## Role in the CR1MS0N platform

C4 is the **C2 validation engine** of the [CR1MS0N continuous adversarial validation platform](https://github.com/CR1MS0N-Operator/veil). Where the platform runs CTEM-style **Validate**-phase emulation, C4 is the tool that materializes the C2 layer on demand — deploying, monitoring, and destroying Mythic/Sliver instances so emulated operations stand up in minutes and tear down cleanly. Deterministic lifecycle and reproducible teardown guarantee every engagement starts from a known-good state.

| Framework | C4's Role |
|-----------|-----------|
| **CTEM** (Continuous Threat Exposure Management) | **Validate** — deploys/manages/destroys the C2 frameworks used for adversarial emulation against validated infrastructure |
| **AEV** (Adversarial Exposure Validation) | Provides the reproducible C2 lifecycle that the platform's emulation and measurement depend on |
| **GRC Engineering** | Scriptable, deterministic lifecycle plus JSON logs produce audit-ready evidence of *what* C2 ran and *for how long* |

**Sibling projects:** [Veil](https://github.com/CR1MS0N-Operator/veil) (validation substrate) · [NightForge](https://github.com/CR1MS0N-Operator/nightforge) (measurement & mobilization) · [Lantern](https://github.com/CR1MS0N-Operator/ACLGuard-Active-Directory-Permission-Auditor) (identity exposure validation).

## Ecosystem

C4 is part of the CR1MS0N-Operator toolset, where each repo owns one facet of the continuous adversarial validation cycle:

- [Veil](https://github.com/CR1MS0N-Operator/veil) — validation substrate: the WireGuard mesh, sensors, and emulation targets C4's C2 operates against
- [NightForge](https://github.com/CR1MS0N-Operator/nightforge) — measurement & mobilization layer: the 10-layer harness and `harnessd` dashboard that turn C4's emulation evidence into decisions
- [Lantern](https://github.com/CR1MS0N-Operator/ACLGuard-Active-Directory-Permission-Auditor) — identity exposure validation: AD permission auditing feeding Discover/Prioritize/Validate

## License

MIT — see [LICENSE](LICENSE).
