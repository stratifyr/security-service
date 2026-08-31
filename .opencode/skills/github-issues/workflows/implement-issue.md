# Workflow C — Implement an issue

Applied when the user asks to *work on* an existing issue (e.g. "lets work on #15", "implement issue #N"). Complements workflow B: B files the ticket, C does the fix. Branch/PR conventions: `references/branch-and-pr.md`.

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
13. **PR — Draft the PR, show it, and confirm:** Draft the full PR title + body exactly per `references/branch-and-pr.md` (body: `closes <issue_link>` plus Summary and Changes sections). Display the draft in a readable code block to the user. Ask: "Push branch and open PR with this title/body?" Do **not** push or create until the user explicitly approves.
14. **PR — Push and open the PR:** On approval, push the branch (`git push -u origin <branch>`) and create the PR per `references/branch-and-pr.md`; return the PR URL.
15. **PR — Ask before closing anything else.** Do not merge the PR, delete the branch, add labels, or tag reviewers unless the user asks.

Key rules: **Per-todo approval + per-todo commit** — one commit per todo, each gated on user confirmation. A blanket "go ahead" concedes the plan, not per-commit checkpoints. **PR needs an explicit push go-ahead** — commits stay local until the PR phase. **Skip nothing implicitly** — if the user wants a step dropped, record it as cancelled in the todo list.