# Branch Protection Rules

> **Instructions:** Apply these rules manually via GitHub UI or API.
>
> Navigate to: `Settings → Branches → Add rule` for `main` branch.

## Recommended Branch Protection Rules for `c4`

### Rule: `main` branch

| Setting | Value | Rationale |
|---------|-------|-----------|
| **Require a pull request before merging** | ✅ Enabled | Enforces PR review discipline |
| **Require approvals** | 1 | Solo operator — self-review with PR discipline |
| **Dismiss stale PR approvals when new commits are pushed** | ✅ Enabled | Ensures latest code is reviewed |
| **Require review from Code Owners** | ❌ Disabled | No CODEOWNERS file configured |
| **Require status checks to pass before merging** | ✅ Enabled | CI must pass |
| **Require branches to be up to date before merging** | ✅ Enabled | Prevents merge conflicts |
| **Require conversation resolution before merging** | ✅ Enabled | All comments addressed |
| **Require signed commits** | ✅ Enabled | Commit authenticity |
| **Require linear history** | ✅ Enabled | Clean git history |
| **Include administrators** | ✅ Enabled | Even admins follow process |
| **Allow force pushes** | ❌ Disabled | Prevent history rewriting |
| **Allow deletions** | ❌ Disabled | Prevent accidental branch deletion |

### Status Checks Required

These checks must pass before merge:

- `commitlint` — Conventional commit format
- `test` — Unit/integration tests (if applicable)
- `lint` — Code linting (if applicable)

### API Application (Alternative)

```bash
# Apply via GitHub API (requires admin token)
gh api repos/CR1MS0N-Operator/c4/branches/main/protection \
  --method PUT \
  --field required_pull_request_reviews='{"required_approving_review_count":1,"dismiss_stale_reviews":true,"require_code_owner_reviews":false}' \
  --field required_status_checks='{"strict":true,"contexts":["commitlint"]}' \
  --field enforce_admins=true \
  --field required_linear_history=true \
  --field allow_force_pushes=false \
  --field allow_deletions=false \
  --field required_conversation_resolution=true
```

## Branch Strategy

```
main (protected) ──────────────────────────────
  │
  ├── feature/*  → PR → main
  ├── fix/*      → PR → main
  ├── docs/*     → PR → main
  └── chore/*    → PR → main
```

### Branch Naming Convention

| Prefix | Purpose | Example |
|--------|---------|---------|
| `feature/` | New features | `feature/add-auth-module` |
| `fix/` | Bug fixes | `fix/login-timeout` |
| `docs/` | Documentation | `docs/update-readme` |
| `chore/` | Maintenance | `chore/update-deps` |
| `security/` | Security fixes | `security/patch-cve-2025-xxxx` |
| `refactor/` | Code refactoring | `refactor/simplify-parser` |
| `test/` | Test additions | `test/add-unit-tests` |
