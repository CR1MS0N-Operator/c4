# C4 — C2 Control Center

C4 is a CLI for managing Command & Control (C2) infrastructure. It deploys, configures, and tears down C2 frameworks (Mythic, Sliver) via GraphQL (Hasura) and Docker Compose.

## Install

```bash
go install github.com/CR1MS0N-Operator/c4@latest
```

Requires Go 1.21+.

## Quick Start

```bash
# Initialize config
c4 init

# Deploy C2 infrastructure
c4 deploy

# Check status
c4 status

# Tear down
c4 destroy
```

## Subcommands

| Command     | Description                                |
|-------------|--------------------------------------------|
| `init`      | Generate default configuration (~/.c4/c4.toml) |
| `deploy`    | Deploy C2 infrastructure via Docker Compose |
| `destroy`   | Tear down deployed infrastructure          |
| `status`    | Show deployment status                     |
| `start`     | Start a stopped service                    |
| `stop`      | Stop a running service                     |
| `listener`  | Manage C2 listeners                        |
| `callback`  | Interact with agent callbacks              |
| `payload`   | Generate or manage payloads                |
| `config`    | View or modify configuration               |
| `detect`    | Run detection rules against infrastructure |

## Requirements

- Docker Engine 24+
- Docker Compose v2
- Hasura GraphQL endpoint (configured via `c4 init`)
- Internet access for container images

## License

MIT
