# Worktrees — dispatch and worker runbook

C4 dispatches work through **treehouse**, a pool of reusable git worktrees under
`~/.treehouse/` (this repo's pool is pinned by [`treehouse.toml`](treehouse.toml):
`root`, `max_trees`, `base_branch = "main"`).

**The shared checkout at `~/Projects/c4` is never a work surface.** Another
session may be using it at any moment. A dispatched worker touches only the
worktree it leased.

## 1. Acquire a worktree

```bash
cd ~/Projects/c4
treehouse get --lease        # prints the absolute worktree path
treehouse status             # slot state for this pool
```

## 2. Create your branch BEFORE touching any file

treehouse hands a worktree over at **detached HEAD**. `git rev-parse
--abbrev-ref HEAD` prints `HEAD`, not a branch name.

A commit made on a detached HEAD is referenced by **no branch**. It is not
reachable by name, `git gc` may reap it once the worktree is gone, and
`treehouse destroy --include-unlanded` discards it outright. This is the most
likely way to lose dispatched work in this repo.

```bash
cd <worktree-path>
git switch -c <bead-id>          # e.g. git switch -c c4-0mn
git rev-parse --abbrev-ref HEAD  # must print <bead-id>, not "HEAD"
```

Do this **before** the first edit, `git add`, or `git commit`.

`treehouse get --branch <bead-id>` creates the branch at acquisition time and
avoids the detached state, but it fails if the branch already exists — so the
sequence above must still be known.

## 3. Work and commit inside the worktree

```bash
cd <worktree-path>
git add <files>
git commit -m "..."
```

Commits land on the branch created in step 2. Never run a mutating git command
against the shared checkout — no `git -C ~/Projects/c4 add`, `commit`,
`checkout`, or `switch`. Staging or committing there clobbers another
session's work.

## 4. Verify the commit is reachable by branch name

```bash
git rev-parse <bead-id>                    # sha of the branch tip
git -C ~/Projects/c4 rev-parse <bead-id>   # same sha, read from the shared repo
```

Branch refs live in the shared repository's git dir, so a committed branch
survives `treehouse return` and any later worktree teardown. Run the second
command before reporting done: if it prints the sha, the work is safe.

## 5. Return the slot

```bash
treehouse return <worktree-path>
```

`treehouse return` hands the slot back to the pool; it does not delete
branches. **Do not** `treehouse destroy` a slot holding unlanded work, and note
that `treehouse prune` only reclaims worktrees whose HEAD is already merged
into the default branch.

## 6. Merge

The supervisor merges the branch into `main` after review:

```bash
git -C ~/Projects/c4 merge --ff-only <bead-id>
```

`main` must be the branch's base for this to stay a fast-forward — rebase the
branch onto `main` first if `main` advanced.

git refuses to merge over untracked files that the branch adds. If the shared
checkout has an untracked file the branch also contains, the merge aborts:

```
error: The following untracked working tree files would be overwritten by merge:
	ARCHITECTURE.md
	SECURITY.md
```

Staging them (`git add`) does **not** help — git then refuses with "Your local
changes to the following files would be overwritten by merge". Remove the
untracked copies first; the merged branch restores them as tracked files:

```bash
cd ~/Projects/c4
rm ARCHITECTURE.md SECURITY.md     # untracked copies; the branch re-adds them
git merge --ff-only <bead-id>
```

Only remove a file after confirming the branch you are merging contains it
(`git ls-tree <bead-id> -- <path>`), so nothing is lost.
