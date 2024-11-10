-- name: CreateUser :one
INSERT INTO users (
    username,
    password,
    email,
    role
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateUserPasswordById :one
UPDATE users
  set password = $2
WHERE id = $1
RETURNING *;

-- name: UpdateUserEmailById :one
UPDATE users
  set email = $2
WHERE id = $1
RETURNING *;

-- name: UpdateUserRoleById :one
UPDATE users
  set email = $2
WHERE id = $1
RETURNING *;

-- name: DeleteUserById :exec
DELETE FROM users
WHERE id = $1;