-- name: UpsertLinks :exec
INSERT INTO
  link(
    id,
    account_id,
    application_id,
    roles,
    grantings,
    oauth_code,
    oauth_expires_at,
    oauth_code_challenge,
    oauth_code_challenge_method,
    created_at,
    updated_at
  )
  SELECT 
    l.id,
    l.account_id,
    l.application_id,
    l.roles,
    l.grantings,
    l.oauth_code,
    l.oauth_expires_at,
    l.oauth_code_challenge,
    l.oauth_code_challenge_method,
    l.created_at,
    l.updated_at
  FROM 
    json_populate_recordset(null::link, $1) as l
ON CONFLICT (id)
DO UPDATE 
  SET
    roles = EXCLUDED.roles,
    grantings = EXCLUDED.grantings,
    oauth_code = EXCLUDED.oauth_code,
    oauth_expires_at = EXCLUDED.oauth_expires_at,
    oauth_code_challenge = EXCLUDED.oauth_code_challenge,
    oauth_code_challenge_method = EXCLUDED.oauth_code_challenge_method,
    updated_at = EXCLUDED.updated_at;