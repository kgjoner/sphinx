-- name: CreateExternalCredential :one
INSERT INTO
  external_credential (
    user_id,
    provider_name,
    provider_subject_id,
    provider_alias,
    last_used_at
  )
VALUES
  (
    (
      SELECT
        internal_id
      FROM
        "user"
      where
        id = $1
    ),
    $2,
    $3,
    $4,
    $5
  );

-- name: UpdateExternalCredential :exec
UPDATE external_credential
SET
  provider_name = $2,
  provider_subject_id = $3,
  provider_alias = $4,
  last_used_at = $5,
  updated_at = NOW()
WHERE
  id = $1;

-- name: RemoveExternalCredential :exec
DELETE FROM external_credential
WHERE
  user_id = (
    SELECT
      internal_id
    FROM
      "user"
    WHERE
      id = $1
  )
  AND provider_name = $2
  AND provider_subject_id = $3;