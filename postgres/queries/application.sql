-- name: CreateApplication :one
INSERT INTO
  application (
    id,
    name,
    grantings
  )
VALUES
  (
    $1,
    $2,
    $3
  )
RETURNING internal_id;

-- name: UpdateApplication :exec
UPDATE
  application
SET
  name = $2,
  grantings = $3
WHERE
  id = $1;

-- name: GetApplicationById :one
SELECT
  *
FROM
  application
WHERE
  id = $1;
