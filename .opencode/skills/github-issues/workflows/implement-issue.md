# Workflow C — Implement an issue

Applied when the user asks to *work on* an existing issue (e.g. "lets work on #15", "implement issue #N"). Complements workflow B: B files the ticket, C does the fix. Prerequisites are in `SKILL.md`; branch/PR conventions live in `references/branch-and-pr.md`.

1. **Plan — Read the issue:** Run `gh issue view <n> --repo <slug>` — capture every list item / checkbox in its "Fix" section; those become the todos.
2. **Plan — Research and verify every claim:** Explore the referenced files; confirm the issue's statements are accurate (dead code really is unreferenced, structs really are triplicated, etc.). Use `go build ./... && go vet ./...` at the start as a clean baseline.
3. **Plan — Present a numbered plan:** One step per issue checklist item, with `file:line` references, mirrored into the `todowrite` todo list. Explicitly call out any decision points the issue leaves open ("delete vs. wire") before proceeding — ask the user and get an answer, rather than assuming.
4. **Plan — Wait for plan approval:** Hold all edits until the plan is approved.
5. **Plan — Create the working branch** from latest development per `references/branch-and-pr.md`. Confirm the branch is active before editing anything.
6. **Build — Make the change for that todo only.** Keep changes scoped — nothing from the next todos leaks in.
7. **Build — Verify:** Run `go build ./... && go vet ./...` (this repo has no tests or lint command).
8. **Build — Show the user what changed:** Summarize the diff, noting any dead-code/import side effects like removed imports.
9. **Build — Ask before committing:** "Anything to alter? Confirm the commit name." Wait for the response. If the user wants changes, apply them, re-verify, and re-confirm.
10. **Build — Commit only the todo's files** with the confirmed message. Never batch in unrelated/untracked files (e.g. a stray `.golangci.yml`). Match the repo's imperative commit style.
11. **Build — Mark the todo `completed`**, then repeat for the next todos.
12. **PR — Final verification:** Run `go build ./... && go vet ./...` and summarize the commit series.
13. **PR — Propose a PR title and confirm the push:** The title must be imperative, matching the issue and commit style. Pushing the branch is required for the PR, so ask explicitly (e.g. "push `15-refactor-metric-cache` and open the PR?") and wait.
14. **PR — Push and open the PR:** On approval, push the branch (`git push -u origin <branch>`) and create the PR per `references/branch-and-pr.md` (body `closes <issue_link>`); return the PR URL.
15. **PR — Ask before closing anything else.** Do not merge the PR, delete the branch, add labels, or tag reviewers unless the user asks.

## Rules

- **Per-todo approval + per-todo commit.** One commit per todo, each gated on the user's confirmation of both the change and the commit message. A blanket "go ahead" concedes the plan, not the per-commit checkpoints.
- **PR needs an explicit push go-ahead.** Commits are local until the PR phase; never `git push` while committing todos.
- **Skip nothing implicitly.** If the issue lists a step the user wants dropped (e.g. "skip the cmd folder"), record that as cancelled in the todo list and stay out of that area.