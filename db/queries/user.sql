-- name: GetUserByEmail :one
SELECT id, person_id, username, email, password_hash, active, created_at, updated_at
FROM users
WHERE email = $1 AND active = true;

-- name: GetUserByID :one
SELECT id, person_id, username, email, password_hash, active, created_at, updated_at
FROM users
WHERE id = $1 AND active = true;
