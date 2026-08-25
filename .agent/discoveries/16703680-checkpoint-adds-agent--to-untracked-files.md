# Checkpoint adds .agent/ to untracked files

**Finding:** When creating the first checkpoint in a new project, the .agent/ directory structure gets created and appears as untracked files in git status after the checkpoint. This is expected behavior since .agent/ is part of project state, but it does change the git status.

*Session: 50c07877-1095-4d19-b8c9-68aeb229dbfc*
*Created: 2026-08-25T14:01:33Z*
