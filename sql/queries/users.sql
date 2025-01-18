-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    now(),
    now(),
    $1,
    $2
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUser :one
UPDATE users SET
    updated_at = now(),
    email = $2,
    hashed_password = $3
WHERE id = $1
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT users.* FROM users
INNER JOIN refresh_tokens ON users.id = refresh_tokens.user_id
WHERE
    refresh_tokens.token = $1
    AND refresh_tokens.expires_at > now()
    AND refresh_tokens.revoked_at IS NULL;
