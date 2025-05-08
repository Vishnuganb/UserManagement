-- name: CreateUser :one
INSERT INTO users (first_name, last_name, email, phone, age, status)
VALUES ($1, $2, $3, $4, $5, $6)
    RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE user_id = $1;

-- name: ListUsers :many
SELECT * FROM users;

-- name: UpdateUser :one
UPDATE users
SET
    first_name = COALESCE(sqlc.narg(first_name), first_name),
    last_name = COALESCE(sqlc.narg(last_name), last_name),
    email = COALESCE(sqlc.narg(email), email),
    phone = COALESCE(sqlc.narg(phone), phone),
    age = COALESCE(sqlc.narg(age), age),
    status = COALESCE(sqlc.narg(status), status),
    updated_at = NOW()
WHERE user_id = $1
    RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE user_id = $1;
