-- name: CreateSession :one
INSERT INTO sessions (
  id,
  user_id,
  token,
  user_agent,
  client_ip,
  expires_at,
  revoked_at,
  created_at,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetSessionByID :one
SELECT *
FROM sessions
WHERE id = $1
LIMIT 1;

-- name: GetSessionByToken :one
SELECT *
FROM sessions
WHERE token = $1
LIMIT 1;

-- name: GetSessionsByUserID :many
SELECT *
FROM sessions
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateSession :one
UPDATE sessions
SET
  token = $2,
  user_agent = $3,
  client_ip = $4,
  expires_at = $5,
  revoked_at = $6,
  updated_at = $7
WHERE id = $1
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;

-- name: DeleteSessionsByUserID :exec
DELETE FROM sessions
WHERE user_id = $1;
