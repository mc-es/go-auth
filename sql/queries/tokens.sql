-- name: CreateToken :one
INSERT INTO tokens (
  id,
  user_id,
  token,
  type,
  expires_at,
  used_at,
  created_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetTokenByID :one
SELECT *
FROM tokens
WHERE id = $1
LIMIT 1;

-- name: GetTokensByUserID :many
SELECT *
FROM tokens
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateToken :one
UPDATE tokens
SET
  token = $2,
  type = $3,
  expires_at = $4,
  used_at = $5
WHERE id = $1
RETURNING *;

-- name: DeleteToken :exec
DELETE FROM tokens
WHERE id = $1;

-- name: DeleteTokensByUserID :exec
DELETE FROM tokens
WHERE user_id = $1;
