# Using C4

[← Project overview](../README.md) · [Architecture](../ARCHITECTURE.md) · [Contributing](../CONTRIBUTING.md)

## Requirements

- The Go toolchain specified in [go.mod](../go.mod) to build or install C4.
- Docker Engine with Docker Compose v2 for Mythic lifecycle commands.
- An existing [Mythic](https://github.com/its-a-feature/Mythic) installation with a Compose project for lifecycle commands, and its Hasura GraphQL endpoint and admin secret for API commands.

C4 does not install Mythic or provision its credentials. Generic exec providers use their own configured local commands and prerequisites.

## Install and configure

```bash
go install github.com/CR1MS0N-Operator/c4@latest
c4 init
```

`c4 init` creates `~/.c4/c4.toml` and `~/.c4/providers/`, then scans for installed frameworks. Use `c4 init --skip-detect` to skip the scan. Initialization will not overwrite an existing configuration.

Edit the generated TOML file before connecting to Mythic:

| Setting under `[mythic]` | Configure it for your installation |
| --- | --- |
| `local_path` | Directory containing the existing Mythic Compose project |
| `host` and `api_port` | Mythic GraphQL endpoint |
| `hasura_secret` | The installation's Hasura admin secret; keep it out of git |
| `ssl` | Use TLS for the connection |
| `cert_fingerprint` or `ca_file` | Trust the server certificate when it is not in the system store |

See [Mythic TLS](#mythic-tls) before sending the admin secret. An alternative configuration path can be selected with `c4 --config /path/to/c4.toml <command>`.

Inspect status after configuring the connection:

```bash
c4 status mythic
```

`c4 status` without a target checks Mythic and configured exec providers. Exec-provider YAML files live in `~/.c4/providers/`; `c4 detect --help` describes discovery and its `--apply` option.

## Command reference

Use `c4 <command> --help` for subcommands and flags.

| Command | Current behavior |
| --- | --- |
| `c4 init` | Generate the default configuration and optionally detect installed frameworks |
| `c4 deploy <c2>` | Bring up an existing Mythic Compose project, or run a configured exec provider's start command |
| `c4 start mythic` | Start the existing Mythic Compose project |
| `c4 stop mythic` | Bring down the Mythic Compose project |
| `c4 destroy <c2>` | Invoke exec-provider teardown; Mythic teardown is not implemented |
| `c4 status [c2]` | Show Mythic and exec-provider status, or filter by name |
| `c4 listener` | Manage Mythic listeners (C2 profiles) |
| `c4 callback` | Manage Mythic callbacks |
| `c4 payload` | List Mythic payloads; generation is not yet implemented |
| `c4 config` | Configuration subcommands: `init`, `set`, and `view` |
| `c4 detect` | Detect installed C2 frameworks |

The standalone `start` and `stop` commands currently accept only Mythic. Exec-provider deployment runs the provider's configured start command; teardown behavior depends on that provider. There is no dedicated Sliver or Havoc provider, and detection does not imply full lifecycle or API support.

**Mythic teardown:** `c4 destroy mythic` currently returns a not-implemented error. Use `c4 stop mythic` to stop the Compose project; it is not a full data-removal operation.

### Logging

| Global flag | Effect |
| --- | --- |
| `--config <path>` | Select a TOML configuration file |
| `--verbose`, `-v` | Enable debug logging |
| `--json` | Emit logs as JSON on stderr; command results are not universally JSON |

## Mythic TLS

C4 sends the Hasura admin secret on every Mythic request, so with `ssl = true` it verifies the server certificate and refuses to connect when the check fails. Mythic ships a self-signed certificate; trust it with exactly one of these `[mythic]` settings in `~/.c4/c4.toml`:

| Setting | Effect |
| --- | --- |
| `cert_fingerprint = "<sha256>"` | Pin the server certificate by SHA-256 fingerprint. Works for certificates without a matching hostname/IP. |
| `ca_file = "/path/ca.pem"` | Verify against this PEM bundle instead of the system trust store. |
| `insecure_skip_verify = true` | Disable verification. Anyone on the network path can read the admin secret. |

Read the fingerprint from the certificate file on the Mythic host, not over the network:

```bash
openssl x509 -in <mythic-cert.pem> -noout -fingerprint -sha256
```

C4 no longer retries on `server_port` when `api_port` fails. Set `server_port_fallback = true` to re-enable the retry; it uses the same scheme and TLS settings as `api_port`.
