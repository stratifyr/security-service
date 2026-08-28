# Workflow A — Audit sweep

Full repo audit producing prioritized tickets one by one. Prerequisites are in `SKILL.md`.

1. **Research first:** Explore the whole codebase (use explore subagents in parallel for large repos). Collect concrete findings with `file:line` references; verify claims by reading the actual code. There is no test suite — evidence comes from reading handlers/services/stores; run `go build ./... && go vet ./...` only if a finding concerns compilability.
2. **Cross-check against open issues:** For every finding, compare against the loaded issue list from the SKILL.md prerequisites (and run targeted `gh issue list --repo <slug> --search "<keyword>"` when unsure). Drop true duplicates from the proposal; for partial overlaps, propose extending the existing ticket instead of a new one.
3. **Propose the full list:** Present a numbered table with priorities so the user sees total scope. Annotate any row that relates to an existing ticket (e.g. "extends #15").
4. **Draft one ticket at a time:** Present title + body in chat per `references/ticket-creation.md`, then ask `Create? (approve / skip / edit)` and WAIT. Apply edits and re-confirm until approved or skipped. Even when the user pre-approves a whole batch or scope up front ("all", "the P1s"), still draft each ticket individually and wait for its own approve / edit / skip response before creating — a blanket approval only sets the candidate list, never waives the gate.
5. **Create approved tickets immediately:** Do so after each approval (or queue if still in plan mode — tell the user creation happens once they exit plan mode). Re-check the live issue list before each creation if much time has passed.
6. **Verify at the end:** Run `gh issue list --state open` and spot-check bodies of any batch created in one command.

## Rules

Ticket format, labels, and creation rules live in `references/ticket-creation.md` — read it before the first draft.