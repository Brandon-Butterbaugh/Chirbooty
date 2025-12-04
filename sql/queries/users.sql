-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(), NOW(), NOW(), $1, $2
)
RETURNING *;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUser :one
SELECT *
FROM users
WHERE email = $1;

-- name: GetUserFromID :one
SELECT *
FROM users
WHERE id = $1;

-- name: UpdateUserPassword :one
UPDATE users
SET updated_at = NOW(), hashed_password = $2
WHERE id = $1
RETURNING *;

-- name: UpdateUserEmail :one
UPDATE users
SET updated_at = NOW(), email = $2
WHERE id = $1
RETURNING *;

-- name: GetRefreshTokenFromUser :one
SELECT refresh_tokens.*
FROM refresh_tokens
JOIN users ON refresh_tokens.user_id = users.id
WHERE users.id = $1;

-- name: UpdateUserRed :one
UPDATE users
SET updated_at = NOW(), is_chirpy_red = TRUE
WHERE id = $1
RETURNING *;