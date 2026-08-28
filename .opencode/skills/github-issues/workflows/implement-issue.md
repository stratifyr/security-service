# Workflow C — Implement an issue

Applied when the user asks to *work on* an existing issue (e.g. "lets work on #15", "implement issue #N"). This complements workflow B: B files the ticket, C does the fix. Prerequisites are in `SKILL.md`. No ticket is created — this workflow commits code and opens a PR.

## 1. Plan phase (before touching any code)

1. **Read the issue:** Run `gh issue view <n> --repo <slug>` — capture every list item / checkbox in its "Fix" section; those become the todos.
2. **Research and verify every claim:** Explore the referenced files; confirm the issue's statements are accurate (dead code really is unreferenced, structs really are triplicated, etc.). Use `go build ./... && go vet ./...` at the start as a clean baseline.
3. **Present a numbered plan:** One step per issue checklist item, with `file:line` references, and mirror it into the `todowrite` todo list. Explicitly call out any decision points the issue leaves open ("delete vs. wire") *before* proceeding — ask the user and get an answer, rather than assuming.
4. **Wait for plan approval:** Hold all edits until the plan is approved.
5. **Create the working branch from latest development:** Run `git fetch`; if the local `development` is stale, `git checkout development && git pull`. Then `git checkout -b <number>-<short-2-4-word-summary>`, e.g. `15-refactor-metric-cache` for issue 15. Confirm the branch is active before editing anything.

## 2. Build phase (one todo at a time)

For **each** todo, in order:

1. **Make the change for that todo only.** Keep changes scoped — nothing from the next todos leaks in.
2. **Verify:** Run `go build ./... && go vet ./...` (this repo has no tests or lint command).
3. **Show the user what changed:** Summarize the diff, noting any dead-code/import side effects like removed imports.
4. **Ask before committing:** "Anything to alter? Confirm the commit name." Wait for the response. If the user wants changes, apply them, re-verify, and re-confirm.
5. **Commit only the todo's files** with the confirmed message. Never batch in unrelated/untracked files (e.g. a stray `.golangci.yml`). Match the repo's imperative commit style.
6. **Mark the todo `completed`**, then repeat for the next todos.

## 3. PR phase (after the last todo)

Once every todo is committed:

1. **Final verification:** Run `go build ./... && go vet ./...` and summarize the commit series.
2. **Propose a PR title and confirm the push:** The title must be imperative, matching the issue and commit style. Pushing the branch is required for the PR, so ask explicitly (e.g. "push `15-refactor-metric-cache` and open the PR?") and wait.
3. **Push and open the PR:** On approval, push the branch (`git push -u origin <branch>`), then create the PR with the body `closes <issue_link>` so GitHub auto-links and closes the issue on merge:
   ```bash
   gh pr create --repo <slug> --title "<confirmed title>" --body "closes https://github.com/<slug>/issues/<n>"
   ```
   Use the full issue URL (or `closes #<n>` on the same repo — either is fine), and return the PR URL.
4. **Ask before closing anything else.** Do not merge the PR, delete the branch, add labels, or tag reviewers unless the user asks.

## Rules (workflow C)

- **Per-todo approval + per-todo commit.** One commit per todo, each gated on the user's confirmation of both the change and the commit message. A blanket "go ahead" concedes the plan, not the per-commit checkpoints.
- **PR needs an explicit push go-ahead.** Commits are local until third phase; never `git push` while committing todos.
- **Skip nothing implicitly.** If the issue lists a step the user wants dropped (e.g. "skip the cmd folder"), record that as cancelled in the todo list and stay out of that area.
- **Verify after the last todo:** Run the final `go build ./... && go vet ./...` and summarize the commit series plainly.