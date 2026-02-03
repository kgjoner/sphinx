-- name: CreateUser :one
INSERT INTO
  "user" (
    id,
    email,
    password,
    phone,
    username,
    document,
    extra_data,
    is_active,
    pending_email,
    pending_phone,
    has_email_been_verified,
    has_phone_been_verified,
    codes
  )
VALUES
  (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13
  );

-- name: UpdateUser :exec
UPDATE "user"
SET
  email = $2,
  password = $3,
  phone = $4,
  username = $5,
  document = $6,
  extra_data = $7,
  is_active = $8,
  pending_email = $9,
  pending_phone = $10,
  has_email_been_verified = $11,
  has_phone_been_verified = $12,
  codes = $13,
  password_updated_at = $14,
  updated_at = $15
WHERE
  id = $1;

-- name: GetUserByID :one
SELECT
  u.id,
  u.email,
  COALESCE(u.phone, ''),
  u.password,
  COALESCE(u.username, ''),
  COALESCE(u.document, ''),
  u.extra_data,
  u.is_active,
  COALESCE(u.pending_email, ''),
  COALESCE(u.pending_phone, ''),
  u.has_email_been_verified,
  u.has_phone_been_verified,
  u.codes,
  u.password_updated_at,
  COALESCE(
    json_agg(
      jsonb_build_object(
        'userID',
        u.id,
        'providerName',
        c.provider_name,
        'providerSubjectID',
        c.provider_subject_id,
        'providerAlias',
        c.provider_alias,
        'lastUsedAt',
        c.last_used_at,
        'createdAt',
        c.created_at,
        'updatedAt',
        c.updated_at
      )
    ) FILTER (
      WHERE
        c.user_id IS NOT NULL
    ),
    '[]'::json
  ) AS external_credentials,
  u.created_at,
  u.updated_at
FROM
  "user" u
  LEFT JOIN external_credential c ON c.user_id = u.internal_id
WHERE
  u.id = $1
GROUP BY
  u.internal_id;

-- name: GetUserByEntry :one
SELECT
  u.id,
  u.email,
  COALESCE(u.phone, ''),
  u.password,
  COALESCE(u.username, ''),
  COALESCE(u.document, ''),
  u.extra_data,
  u.is_active,
  COALESCE(u.pending_email, ''),
  COALESCE(u.pending_phone, ''),
  u.has_email_been_verified,
  u.has_phone_been_verified,
  u.codes,
  u.password_updated_at,
  COALESCE(
    json_agg(
      jsonb_build_object(
        'userID',
        u.id,
        'providerName',
        c.provider_name,
        'providerSubjectID',
        c.provider_subject_id,
        'providerAlias',
        c.provider_alias,
        'lastUsedAt',
        c.last_used_at,
        'createdAt',
        c.created_at,
        'updatedAt',
        c.updated_at
      )
    ) FILTER (
      WHERE
        c.user_id IS NOT NULL
    ),
    '[]'::json
  ) AS external_credentials,
  u.created_at,
  u.updated_at
FROM
  "user" u
  LEFT JOIN external_credential c ON c.user_id = u.internal_id
WHERE
  u.email = $1
  OR u.phone = $1
  OR u.username = $1
  OR u.document = $1
GROUP BY
  u.internal_id;

-- name: GetUserByExternalCredential :one
SELECT
  u.id,
  u.email,
  COALESCE(u.phone, ''),
  u.password,
  COALESCE(u.username, ''),
  COALESCE(u.document, ''),
  u.extra_data,
  u.is_active,
  COALESCE(u.pending_email, ''),
  COALESCE(u.pending_phone, ''),
  u.has_email_been_verified,
  u.has_phone_been_verified,
  u.codes,
  u.password_updated_at,
  COALESCE(
    json_agg(
      jsonb_build_object(
        'userID',
        u.id,
        'providerName',
        c.provider_name,
        'providerSubjectID',
        c.provider_subject_id,
        'providerAlias',
        c.provider_alias,
        'lastUsedAt',
        c.last_used_at,
        'createdAt',
        c.created_at,
        'updatedAt',
        c.updated_at
      )
    ) FILTER (
      WHERE
        c.user_id IS NOT NULL
    ),
    '[]'::json
  ) AS external_credentials,
  u.created_at,
  u.updated_at
FROM
  "user" u
  LEFT JOIN external_credential c ON c.user_id = u.internal_id
WHERE
  c.provider_name = $1
  AND c.provider_subject_id = $2
GROUP BY
  u.internal_id;