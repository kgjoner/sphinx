-- name: CreateApplication :one
INSERT INTO
  application (
    id,
    name,
    possible_roles,
    secret,
    allowed_redirect_uris,
    brand
  )
VALUES
  (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
  )
RETURNING internal_id;

-- name: UpdateApplication :exec
UPDATE
  application
SET
  name = $2,
  possible_roles = $3,
  allowed_redirect_uris = $4,
  brand = $5
WHERE
  id = $1;

-- name: GetApplicationById :one
SELECT
  internal_id,
  id,
  name,
  possible_roles,
  secret,
  allowed_redirect_uris,
  brand,
  created_at,
  updated_at
FROM
  application
WHERE
  id = $1;
