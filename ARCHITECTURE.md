# C4 Architecture

C4 is a CLI control center for deploying, managing, and destroying C2 frameworks (Mythic, Sliver) via Docker Compose.

## Stack

- **Language:** Go 1.26
- **CLI:** Cobra
- **Infrastructure:** Docker Compose, Podman Quadlet
- **Backend:** GraphQL (Hasura) for state management

## Structure

```
cmd/            CLI commands (Cobra)
├── root.go     Root command + global flags
├── deploy.go   Deploy a C2 instance
├── start.go    Start a stopped instance
├── stop.go     Stop a running instance
├── destroy.go  Tear down an instance
├── detect.go   Scan host for existing C2 frameworks
├── listener.go Manage listeners
├── callback.go Manage callbacks
├── payload.go  Manage payloads
├── config.go   Configuration management
├── init.go     Project initialization
└── status.go   Show instance status

pkg/            Core libraries
├── provider/   C2 provider abstraction interface
├── mythic/     Mythic C2 provider implementation
├── execprovider/ Generic exec-based provider
├── docker/     Docker Compose orchestration
├── graphql/    Hasura GraphQL client
├── config/     Configuration loading/saving
├── detect/     Framework detection logic
├── template/   Container template management
└── c4lib/      Shared utilities

resources/      Docker Compose templates and configs
```

## Provider abstraction

```
pkg/provider/interface.go
    ├── MythicProvider    (pkg/mythic/)
    └── ExecProvider      (pkg/execprovider/)
```

New C2 frameworks implement the `provider.Provider` interface without touching CLI code.

## Data flow

```
User → c4 CLI (cmd/) → Provider interface (pkg/provider/)
    → Docker Compose (pkg/docker/) → C2 containers
    → GraphQL state (pkg/graphql/) → Hasura backend
```
