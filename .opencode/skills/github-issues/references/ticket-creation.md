# Ticket format and creation rules (workflows A and B)

Shared by workflows A and B. Workflow C (`workflows/implement-issue.md`) does **not** create a ticket — it commits code and opens a PR, so none of this applies to it.

## Format

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

## Rules

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