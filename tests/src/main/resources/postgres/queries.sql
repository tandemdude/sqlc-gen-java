-- name: CreateUser :exec
INSERT INTO users(user_id, username, email) VALUES ($1, $2, $3);

-- name: GetUser :one
SELECT * FROM users WHERE user_id = $1;

-- name: GetUserDup :one
SELECT * FROM users WHERE user_id IS NOT NULL;

-- name: ListUsers :many
SELECT * FROM users;

-- name: GetMessage :one
SELECT * FROM messages WHERE message_id = $1;

-- name: CreateMessage :one
INSERT INTO messages(chat_id, user_id, content, attachments)
VALUES ($1, $2, $3, $4)
RETURNING message_id;
