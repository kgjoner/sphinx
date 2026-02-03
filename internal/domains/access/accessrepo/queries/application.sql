-- name: CreateApplication :exec
INSERT INTO
  application (
    id,
    name,
    possible_roles,
    secret,
    allowed_redirect_uris
  )
VALUES
  ($1, $2, $3, $4, $5);

-- name: UpdateApplication :exec
UPDATE application
SET
  name = $2,
  possible_roles = $3,
  allowed_redirect_uris = $4
WHERE
  id = $1;

-- name: GetApplicationByID :one
SELECT
  id,
  name,
  possible_roles,
  secret,
  allowed_redirect_uris,
  created_at,
  updated_at
FROM
  application
WHERE
  id = $1;
