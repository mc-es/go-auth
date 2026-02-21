-- name: CreateUser :one
INSERT INTO users (
  id,
  username,
  email,
  password,
  first_name,
  last_name,
  role,
  status,
  verified_at,
  created_at,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1
LIMIT 1;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1
LIMIT 1;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1
LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET
  username = $2,
  email = $3,
  password = $4,
  first_name = $5,
  last_name = $6,
  role = $7,
  status = $8,
  verified_at = $9,
  updated_at = $10
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: ExistsByUsername :one
SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);

-- name: ExistsByEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);
