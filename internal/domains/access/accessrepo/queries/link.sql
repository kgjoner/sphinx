-- name: UpsertLinks :exec
INSERT INTO
  link(
    id,
    user_id,
    application_id,
    roles,
    has_consent,
    created_at,
    updated_at
  )
  SELECT 
    l.id,
    (SELECT internal_id FROM "user" WHERE id = l.user_id),
    (SELECT internal_id FROM application WHERE id = l.application_id),
    l.roles,
    l.has_consent,
    l.created_at,
    l.updated_at
  FROM 
    json_populate_recordset(null::link, $1) as l
ON CONFLICT (id)
DO UPDATE 
  SET
    roles = EXCLUDED.roles,
    has_consent = EXCLUDED.has_consent,
    updated_at = EXCLUDED.updated_at;

-- name: GetUserLink :one
SELECT
  l.id,
  u.id AS user_id,
  jsonb_build_object(
    'id', app.id,
    'name', app.name,
    'possible_roles', app.possible_roles,
    'secret':  app.secret,
    'allowed_redirect_uris', app.allowed_redirect_uris,
    'created_at', app.created_at,
    'updated_at', app.updated_at
  ) AS application,
  l.roles,
  l.has_consent,
  l.created_at,
  l.updated_at
FROM
  link l
  LEFT JOIN "user" u ON l.user_id = u.internal_id
  LEFT JOIN application app ON l.application_id = app.internal_id
WHERE
  u.id = $1
  AND app.id = $2;