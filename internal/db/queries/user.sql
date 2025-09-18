-- name: InsertUser :one
INSERT INTO
  users (email, password, initial_ip_address)
VALUES
  (?, ?, ?) RETURNING id;

-- name: GetUserByEmail :one
SELECT
  email,
  password,
  id
FROM
  users
WHERE
  email = ?;

-- name: GetUserByID :one
SELECT
  email,
  password,
  id
FROM
  users
WHERE
  id = ?;

-- name: DeleteUserByID :exec
DELETE FROM users
WHERE
  id = ?;

-- name: UpdateUserPassword :exec
UPDATE users
SET
  password = ?
WHERE
  id = ?;
