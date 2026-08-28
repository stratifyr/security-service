---
name: github-issues
description: Use when the user wants GitHub tickets/issues created from codebase findings ("check for improvements", "audit the repo", "create github tickets", "plan a TODO list", "improve <flow/component>", "file this as an issue", or when a bug/flaw is discovered mid-task), OR when the user wants to work on / close an existing issue ("work on issue", "implement issue #15", "fix issue #N", "lets work on <url>"). Covers three workflows a) full repo audit producing prioritized tickets one by one, b) single ticket for one specific finding, c) implementing a single open issue with per-todo commits and a closing PR.
---

# GitHub Issues workflow

Issue-related work end to end. Three workflows share the prerequisites below; pick by scope, then **read the matching workflow file in full** — the steps and rules live there so only the relevant context is loaded. Files are relative to this skill's base directory.

## Pick the workflow

- **A. Audit sweep** → read `workflows/audit-sweep.md`. Triggered by repo-wide improvement requests ("check for improvements", "plan a TODO list with priorities", "create tickets for everything wrong").
- **B. Single ticket** → read `workflows/single-ticket.md`. Triggered when the user names one endpoint/component to improve, or a concrete flaw surfaces mid-task. Exactly one ticket comes out of it.
- **C. Implement an issue** → read `workflows/implement-issue.md`. Triggered when the user points at an existing open issue to close ("lets work on #15"). Plan + per-todo commits, no ticket is created.

## Prerequisites (all workflows)

1. Run `gh auth status`. If not authenticated, ask the user to run `gh auth login` before proceeding.
2. Resolve the repo slug from `git remote -v` (e.g. `stratifyr/security-service`) instead of assuming it.
3. **Load existing issues before drafting anything** (workflows A/B, to avoid duplicates; workflow C instead reads the specific issue with `gh issue view <n>`):
   ```bash
   gh issue list --repo <slug> --state open --limit 100 --json number,title,labels
   ```
   Keep this list in mind for the whole session so nothing gets filed twice.

## Shared references

- **Ticket format, labels, and creation rules** (workflows A and B): read `references/ticket-creation.md` before drafting the first ticket. Workflow C creates no ticket and does not need it.
- **Branch and PR conventions** (workflow C only): read `references/branch-and-pr.md` before creating the branch or the closing PR. Workflows A and B create neither.