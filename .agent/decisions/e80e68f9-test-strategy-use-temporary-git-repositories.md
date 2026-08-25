# Test strategy: Use temporary Git repositories

**Decision:** Integration tests should use temporary Git repositories to avoid interfering with the actual repository and allow for destructive testing

**Rationale:** Git operations modify repository state, so tests need isolated environments to avoid conflicts and ensure reproducibility


*Session: 50c07877-1095-4d19-b8c9-68aeb229dbfc*
*Created: 2026-08-25T13:54:46Z*
