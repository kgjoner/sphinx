-- name: UpsertLinks :exec
INSERT INTO
  link(
    id,
    account_id,
    application_id,
    roles,
    has_consent,
    created_at,
    updated_at
  )
  SELECT 
    l.id,
    l.account_id,
    l.application_id,
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