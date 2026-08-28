# Branch and PR conventions (workflow C)

Shared conventions for branch creation and the closing PR. Used by workflow C (`workflows/implement-issue.md`); workflows A and B create no branches or PRs.

## Branch

Created from latest `development` before any code changes:

- **Name:** `<number>-<short-2-4-word-summary>`, e.g. `15-refactor-metric-cache` for issue 15.
- **Refresh `development`:** `git fetch`; `git checkout development && git pull` if the local copy is stale.
- **Create it:** `git checkout -b <number>-<short-2-4-word-summary>`.
- Confirm the branch is active before editing anything.

## PR

- **Title:** imperative, matching the issue and commit style.
- **Body:** `closes <issue_link>` (full URL or `closes #<n>` on the same repo) so GitHub auto-links and closes the issue on merge.
- **Push then create:** pushing requires an explicit go-ahead; then run:
  ```bash
  gh pr create --repo <slug> --title "<confirmed title>" --body "closes https://github.com/<slug>/issues/<n>"
  ```
- Return the PR URL. Do not merge the PR, delete the branch, add labels, or tag reviewers unless the user asks.