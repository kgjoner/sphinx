-- name: GetClient :one
SELECT
  id,
  id as audience_id,
  secret,
  name,
  allowed_redirect_uris
FROM
  application
WHERE
  id = $1;