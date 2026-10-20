/**
 * Commitlint configuration for the "Lint Commits" workflow.
 *
 * The workflow installs @commitlint/config-conventional but supplied no config,
 * so commitlint loaded zero rules and exited 9 with [empty-rules] on every push.
 * Merge commits are ignored so that PR merges do not fail the conventional-commit
 * gate.
 */
module.exports = {
  extends: ['@commitlint/config-conventional'],
  ignores: [(message) => /^Merge (pull request|branch|remote-tracking)/.test(message)],
};
