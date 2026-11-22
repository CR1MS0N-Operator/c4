# C4 Architecture

C4 is a CLI control center for Mythic and generic exec providers. Mythic uses Docker Compose for lifecycle operations and Hasura GraphQL for API operations. There is no dedicated Sliver provider, and Mythic destruction is not yet implemented.

[Project overview](README.md) · [Usage guide](docs/USAGE.md) · [Contributing](CONTRIBUTING.md)

## Stack

- **Language:** Go (see [go.mod](go.mod) for the required toolchain)
- **CLI:** Cobra
- **Infrastructure:** Docker Compose
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

The `provider.Provider` interface defines the shared contract. Current CLI lifecycle commands dispatch explicitly to Mythic or configured exec providers; a new native provider also needs CLI wiring. The standalone `start` and `stop` commands currently support only Mythic.

## Data flow

```
User → c4 CLI (cmd/) → Provider interface (pkg/provider/)
    → Docker Compose (pkg/docker/) → C2 containers
    → GraphQL state (pkg/graphql/) → Hasura backend
```
