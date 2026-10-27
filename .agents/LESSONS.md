# Agent Lessons Learned

This file records operational lessons, invariants, and failure modes observed during development.

## Go Operational Lessons
- Tenant Isolation: Every query on `notes` must include `WHERE owner_id = $X` to enforce multi-tenant isolation.
- Dependency Injection: Always inject `*db.Queries` into handler structs. Never instantiate global database pools.
- SQLC Workflow: Modify `sql/schema.sql` and `sql/queries.sql`, then run `make generate` to regenerate `internal/db/`.