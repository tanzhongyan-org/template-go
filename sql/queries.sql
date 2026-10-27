-- name: ListNotes :many
SELECT id, owner_id, title, body, created_at, updated_at
FROM notes
WHERE owner_id = $1
ORDER BY created_at DESC;

-- name: GetNote :one
SELECT id, owner_id, title, body, created_at, updated_at
FROM notes
WHERE id = $1 AND owner_id = $2
LIMIT 1;

-- name: CreateNote :one
INSERT INTO notes (id, owner_id, title, body, created_at, updated_at)
VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, owner_id, title, body, created_at, updated_at;

-- name: UpdateNote :one
UPDATE notes
SET title = COALESCE(sqlc.narg('title'), title),
    body = COALESCE(sqlc.narg('body'), body),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND owner_id = $2
RETURNING id, owner_id, title, body, created_at, updated_at;

-- name: DeleteNote :execrows
DELETE FROM notes
WHERE id = $1 AND owner_id = $2;