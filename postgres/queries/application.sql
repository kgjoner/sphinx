-- name: CreateApplication :one
INSERT INTO
  application (
    id,
    name,
    grantings,
    secret,
    allowed_redirect_uris
  )
VALUES
  (
    $1,
    $2,
    $3,
    $4,
    $5
  )
RETURNING internal_id;

-- name: UpdateApplication :exec
UPDATE
  application
SET
  name = $2,
  grantings = $3,
  allowed_redirect_uris = $4
WHERE
  id = $1;

-- name: GetApplicationById :one
SELECT
  *
FROM
  application
WHERE
  id = $1;
