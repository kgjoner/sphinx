-- name: CreateExternalCredential :one
INSERT INTO
  external_credential (
    user_id,
    provider,
    subject_id,
    has_consent,
    last_used_at,
  ) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
  ) 
  RETURNING id;

-- name: UpdateExternalCredential :exec
UPDATE
  external_credential
SET
  provider = $2,
  subject_id = $3,
  has_consent = $4,
  last_used_at = $5,
  updated_at = NOW()
WHERE
  id = $1;

-- name: RemoveExternalCredential :exec
DELETE ec 
  FROM
    external_credential ec
    JOIN "user" u ON ec.user_id = u.internal_id
  WHERE
    u.id = $1 AND
    ec.provider = $2 AND
    ec.subject_id = $3;