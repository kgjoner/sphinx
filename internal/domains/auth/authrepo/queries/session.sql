-- name: CreateSession :exec
INSERT INTO
  session (
    id,
    user_id,
    application_id,
    refresh_token,
    elapsed_minutes_between_refreshes,
    refreshes_count,
    device,
    ip,
    is_active
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
    $4,
    $5,
    $6,
    $7,
    $8,
    $9
  );

-- name: UpdateSession :exec
UPDATE session
SET
  refresh_token = $2,
  refreshed_at = $3,
  elapsed_minutes_between_refreshes = $4,
  refreshes_count = $5,
  is_active = $6,
  terminated_at = $7,
  updated_at = $8
WHERE
  id = $1;

-- name: GetSessionByID :one
SELECT
  s.id,
  u.id as subject_id,
  u.email as subject_email,
  COALESCE(
    u.extra_data ->> 'name',
    u.username,
    SUBSTRING(u.email, 1, POSITION('@' IN u.email) - 1)
  ) as subject_name,
  a.id as audience_id,
  l.roles as roles,
  s.ip,
  s.device,
  s.refresh_token,
  s.refreshed_at,
  s.elapsed_minutes_between_refreshes,
  s.refreshes_count,
  s.is_active,
  s.terminated_at,
  s.created_at,
  s.updated_at
FROM
  session s
  JOIN "user" u ON s.user_id = u.internal_id
  JOIN application a ON s.application_id = a.internal_id
  JOIN link l ON l.user_id = u.internal_id
  AND l.application_id = a.internal_id
WHERE
  s.id = $1;

-- name: TerminateAllSubjectSessions :exec
UPDATE session
SET
  is_active = FALSE,
  terminated_at = NOW(),
  updated_at = NOW()
WHERE
  user_id = (
    SELECT
      internal_id
    FROM
      "user"
    WHERE
      id = $1
  );