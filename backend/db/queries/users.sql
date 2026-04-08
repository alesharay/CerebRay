-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByOIDCSubject :one
SELECT * FROM users WHERE oidc_subject = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (oidc_subject, email, name, avatar_url)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateUser :one
UPDATE users SET
    email = $2,
    name = $3,
    avatar_url = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserPreferences :one
UPDATE users SET
    preferences = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
