# .agent/ gitignore policy differs between source repo and user repos

**Finding:** The .gitignore in the Agent Ledger source repo excludes .agent/ because it contains per-developer session state for working on Agent Ledger itself. In USER projects, .agent/ should be committed to Git — it is the persistent ledger. This distinction needs to be documented clearly.

*Session: 4417de2f-d064-4f67-b491-9a71f20fa80b*
*Created: 2026-08-25T16:43:47Z*
