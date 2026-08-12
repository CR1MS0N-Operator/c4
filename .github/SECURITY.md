# Security Policy

## Reporting a Vulnerability

C4 is a red-team / offensive-security tool. If you find a security
vulnerability — in the CLI itself, its container templates, or the way it
handles configuration and secrets — please report it **privately** so it can be
fixed before disclosure:

- Use GitHub's private vulnerability reporting: **Security → Report a
  vulnerability** on the [c4 repository](https://github.com/CR1MS0N-Operator/c4/security).
- Do **not** open a public issue for undisclosed vulnerabilities.

Please include:

- A description of the issue and its impact
- Steps to reproduce
- Affected version(s) (see `c4 --version`)
- Any suggested fix, if you have one

## Supported Versions

| Version | Supported          |
|---------|--------------------|
| latest  | ✅ Supported       |
| < 0.1.0 | ❌ Unsupported     |

## Responsible Use

C4 is intended for use against infrastructure you own or are authorized to
test. Unauthorized deployment of C2 infrastructure may violate applicable laws
and policies; operators are responsible for ensuring authorization before use.
