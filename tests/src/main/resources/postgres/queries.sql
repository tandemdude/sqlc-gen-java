-- name: CreateUser :exec
INSERT INTO users(user_id, username, email) VALUES ($1, $2, $3);

-- name: GetUser :one
SELECT * FROM users WHERE user_id = $1;

-- name: ListUsers :many
SELECT * FROM users;
