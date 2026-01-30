-- name: CreateLink :exec
INSERT INTO
  link (
    id,
    user_id,
    application_id,
    roles,
    has_consent
  )
VALUES
  (
    $1,
    (
      SELECT
        internal_id
      FROM
        "user"
      WHERE
        id = $2
    ),
    (
      SELECT
        internal_id
      FROM
        application
      WHERE
        id = $3
    ),
    $4::text[],
    $5
  );

-- name: UpdateLink :exec
UPDATE
  link
SET
  roles = $2,
  has_consent = $3,
  updated_at = $4
WHERE
  id = $1;

-- name: GetUserLink :one
SELECT
  l.id,
  u.id AS user_id,
  jsonb_build_object(
    'id',
    app.id,
    'name',
    app.name,
    'possible_roles',
    app.possible_roles,
    'secret',
    app.secret,
    'allowed_redirect_uris',
    app.allowed_redirect_uris,
    'created_at',
    app.created_at,
    'updated_at',
    app.updated_at
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