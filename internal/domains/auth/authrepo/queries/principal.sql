-- name: GetPrincipal :one
SELECT
  u.id,
  'user' as kind,
  u.password,
  u.email,
  COALESCE(
    u.extra_data ->> 'name',
    u.username,
    SUBSTRING(u.email, 1, POSITION('@' IN u.email) - 1)
  ) as "name",
  a.id as audience_id,
  l.roles,
  l.has_consent,
  u.is_active,
  json_agg(
    jsonb_build_object(
      'provider_name',
      ec.provider_name,
      'provider_subject_id',
      ec.provider_subject_id
    )
  ) FILTER (
    WHERE
      ec.user_id IS NOT NULL
  ) AS external_credentials
FROM
  "user" u
  JOIN link l ON u.internal_id = l.user_id
  JOIN application a ON l.application_id = a.internal_id
  LEFT JOIN external_credential ec ON ec.user_id = u.internal_id
WHERE
  u.id = $1
  AND a.id = $2
GROUP BY u.internal_id, a.internal_id, l.roles, l.has_consent;

-- name: GetPrincipalByEntry :one
SELECT
  u.id,
  'user' as kind,
  u.password,
  u.email,
  COALESCE(
    u.extra_data ->> 'name',
    u.username,
    SUBSTRING(u.email, 1, POSITION('@' IN u.email) - 1)
  ) as "name",
  a.id as audience_id,
  l.roles,
  l.has_consent,
  u.is_active,
  json_agg(
    jsonb_build_object(
      'provider_name',
      ec.provider_name,
      'provider_subject_id',
      ec.provider_subject_id
    )
  ) FILTER (
    WHERE
      ec.user_id IS NOT NULL
  ) AS external_credentials
FROM
  "user" u
  JOIN link l ON u.internal_id = l.user_id
  JOIN application a ON l.application_id = a.internal_id
  LEFT JOIN external_credential ec ON ec.user_id = u.internal_id
WHERE
  (
    u.email = $1
    OR u.phone = $1
    OR u.username = $1
    OR u.document = $1
  )
  AND a.id = $2
GROUP BY u.internal_id, a.internal_id, l.roles, l.has_consent;

-- name: GetPrincipalByExtSubID :one
SELECT
  u.id,
  'user' as kind,
  u.password,
  u.email,
  COALESCE(
    u.extra_data ->> 'name',
    u.username,
    SUBSTRING(u.email, 1, POSITION('@' IN u.email) - 1)
  ) as "name",
  a.id as audience_id,
  l.roles,
  l.has_consent,
  u.is_active,
  json_agg(
    jsonb_build_object(
      'provider_name',
      ec.provider_name,
      'provider_subject_id',
      ec.provider_subject_id
    )
  ) FILTER (
    WHERE
      ec.user_id IS NOT NULL
  ) AS external_credentials
FROM
  "user" u
  JOIN link l ON u.internal_id = l.user_id
  JOIN application a ON l.application_id = a.internal_id
  LEFT JOIN external_credential ec ON ec.user_id = u.internal_id
WHERE
  ec.provider_name = $1
  AND ec.provider_subject_id = $2
  AND a.id = $3
GROUP BY u.internal_id, a.internal_id, l.roles, l.has_consent;