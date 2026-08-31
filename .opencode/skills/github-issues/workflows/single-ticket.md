# Workflow B — Single ticket

One specific finding → exactly one ticket.

1. **Investigate the specific area:** Read the named endpoint/component plus its direct dependencies; pin down exact `file:line` evidence and impact. Do NOT expand into a full audit.
2. **Dedup check (mandatory):** Search open issues for the component keywords (`gh issue list --repo <slug> --search "<component or symptom>"`). If one already covers it, do not create a duplicate — show the existing ticket to the user and offer to extend its body (`gh issue edit --body-file -`), bump its label/priority, or close-and-supersede as they prefer.
3. **Draft exactly one ticket:** Follow the format in `references/ticket-creation.md` and present it in chat.
4. **Ask before creating:** Present `Create? (approve / skip / edit)` and wait. If the user instead wants the fix implemented right away, do that — and still offer to file the ticket so the change is tracked.
5. **Create on approval:** Verify with `gh issue view <n>`.

Ticket format, labels, and creation rules: `references/ticket-creation.md`.