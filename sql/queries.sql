-- name: GetUser :one
SELECT id, name, email, created_at, updated_at FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT id, name, email, created_at, updated_at FROM users
WHERE email = $1 LIMIT 1;

-- name: ListUsers :many
SELECT id, name, email, created_at, updated_at FROM users
ORDER BY id;

-- name: CreateUser :one
INSERT INTO users (name, email)
VALUES ($1, $2)
RETURNING id, name, email, created_at, updated_at;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
