# Lessons

Pain-point rules load only on failure or a match. Add one row per new failure before you close a task. Cap: 200 lines.

## Rules

| # | Rule | Owner | Detection |
| --- | --- | --- | --- |
| 1 | Every query on `notes` filters by `owner_id`. | `sql/queries.sql` | Review; cross-owner test in `cmd/server/notes_test.go`. |
| 2 | Inject `*db.Queries` into handlers. Never create a global pool. | `internal/handler` | Review. |
| 3 | Change `sql/schema.sql` or `sql/queries.sql`, then run `make generate`. | `internal/db` | `make generate-check` (sqlc diff). |
| 4 | Send responses through `writeJSON` and `writeError`, not `http.Error` or a bare encoder. | `internal/handler/respond.go` | golangci-lint `dupl`; review. |
| 5 | Identity comes only from `requireOwner`; never fall back to a default owner. | `internal/handler/notes.go` | 401 tests in `cmd/server/main_test.go`. |

## Incidents

| Date | Task | Mistake | Fix | Detection |
| ---- | ---- | ------- | --- | --------- |
