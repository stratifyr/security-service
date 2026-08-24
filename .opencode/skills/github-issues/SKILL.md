---
name: github-issues
description: Use when the user wants GitHub tickets/issues created from codebase findings — triggers include "check for improvements", "audit the repo", "create github tickets", "plan a TODO list", as well as "improve <flow/component>", "file this as an issue", or when a bug/flaw is discovered mid-task. Covers two workflows a) full repo audit producing prioritized tickets one by one, b) single ticket for one specific finding.
---

# GitHub Issues workflow

Create well-researched GitHub issues from repo findings. Two workflows share the same conventions; pick based on scope.

## Pick the workflow

- **A. Audit sweep** — user asks to check/find improvements across the repo ("check for improvements", "plan a TODO list with priorities", "create tickets for everything wrong").
- **B. Single ticket** — one specific thing: the user names an endpoint/component to improve, or a concrete flaw surfaces while doing other work. Exactly one ticket comes out of it.

## Prerequisites (both)

1. Run `gh auth status`. If not authenticated, ask the user to run `gh auth login` before proceeding.
2. Resolve the repo slug from `git remote -v` (e.g. `stratifyr/security-service`) instead of assuming it.
3. **Load existing issues before drafting anything**:
   ```bash
   gh issue list --repo <slug> --state open --limit 100 --json number,title,labels
   ```
   Keep this list in mind for the whole session so nothing gets filed twice.

---

## Workflow A — Audit sweep

1. **Research first.** Explore the whole codebase (use explore subagents in parallel for large repos). Collect concrete findings with `file:line` references; verify claims by reading the actual code. There is no test suite — evidence comes from reading handlers/services/stores; run `go build ./... && go vet ./...` only if a finding concerns compilability.
2. **Cross-check against open issues.** For every finding, compare against the loaded issue list (and run targeted `gh issue list --repo <slug> --search "<keyword>"` when unsure). Drop true duplicates from the proposal; for partial overlaps, propose extending the existing ticket instead of a new one.
3. **Propose the full list** as a numbered table with priorities so the user sees total scope. Annotate any row that relates to an existing ticket (e.g. "extends #15").
4. **Draft one ticket at a time**: title + body in chat, then ask `Create? (approve / skip / edit)` and WAIT. Apply edits and re-confirm until approved or skipped. Even when the user pre-approves a whole batch or scope up front ("all", "the P1s"), still draft each ticket individually and wait for its own approve / edit / skip response before creating — a blanket approval only sets the candidate list, never waives the gate.
5. **Create approved tickets immediately** after each approval (or queue if still in plan mode — tell the user creation happens once they exit plan mode). Re-check the live issue list before each creation if much time has passed.
6. **Verify at the end**: `gh issue list --state open` and spot-check bodies of any batch created in one command.

## Workflow B — Single ticket

1. **Investigate the specific area**: read the named endpoint/component plus its direct dependencies; pin down exact `file:line` evidence and impact. Do NOT expand into a full audit.
2. **Dedup check (mandatory)**: search open issues for the component keywords (`gh issue list --repo <slug> --search "<component or symptom>"`). If one already covers it, do not create a duplicate — show the existing ticket to the user and offer to extend its body (`gh issue edit --body-file -`), bump its label/priority, or close-and-supersede as they prefer.
3. **Draft exactly one ticket** (same format as below) and present it in chat.
4. Ask `Create? (approve / skip / edit)` and wait. If the user instead wants the fix implemented right away, do that — and still offer to file the ticket so the change is tracked.
5. Create on approval; verify with `gh issue view <n>`.

---

## Ticket format (both)

```markdown
**Title:** <imperative summary, no priority prefix>

**Labels:** <type>,<priority>

**Body:**
> ## Problem
> What is wrong, with `path/to/file.go:<line>` references and short code snippets.
>
> ## Fix
> - [ ] Concrete checklist items
```

Body goes to GitHub verbatim via stdin:

```bash
gh issue create --repo <slug> --title "<title>" --label <type>,<priority> --body-file - <<'EOF'
## Problem
...
## Fix
- [ ] ...
EOF
```

## Rules (both)

- **Never create duplicates.** Every draft must state (silently, at minimum) that no open issue already covers it. Overlapping-but-not-identical findings default to extending the existing ticket, not filing a new one — unless the user explicitly wants a separate ticket.
- **No `[P0]`/`[P1]` in titles.** Priority lives in labels only.
- **Labels:** type = `bug` or `enhancement` or `feature`; priority = `P0`, `P1`, `P2`, `P3`.
  Create missing labels once before the first issue (idempotent):
  ```bash
  gh label create "P0" --color b60205 --description "Critical: security or crash bugs" || true
  gh label create "P1" --color d93f0b --description "High: reliability / API correctness" || true
  gh label create "P2" --color fbca04 --description "Medium: dedup / maintainability" || true
  gh label create "P3" --color 0e8a16 --description "Low: polish / hygiene" || true
  ```
  Map findings: security/crash/data-loss → `bug,P0`; broken endpoints/reliability/data-integrity → `bug,P1`; refactors/dedup → `enhancement`; hygiene/polish → lower priority.
- **ONE heredoc per shell command.** Never chain two `--body-file - <<'EOF'` commands in a single Bash call — both heredocs collapse into the first command's stdin and produce merged/empty bodies. Multiple creations = multiple sequential Bash calls, each with exactly one heredoc.
- **Quote titles safely:** single-quote unless the title contains `'`; prefer rewording over escaping.
- **Respect approval gates.** Never create a ticket the user has not approved in the current conversation.
