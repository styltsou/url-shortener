-- name: CreateUser :one
INSERT INTO users (clerk_id, email, avatar_url)
VALUES ($1, $2, sqlc.narg('avatar_url'))
RETURNING *;

-- name: GetUserByClerkID :one
SELECT * FROM users
WHERE clerk_id = $1
LIMIT 1;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;