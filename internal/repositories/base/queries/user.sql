-- name: CreateUser :one
INSERT INTO
  user (
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
  )
RETURNING internal_id;

-- name: UpdateUser :exec
UPDATE
  user
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
WITH la AS (
  SELECT
    l.*,
    json_agg(app.*)->0 application
  FROM
    link l
    JOIN application app ON app.internal_id = l.application_id
  GROUP BY
    l.internal_id
), sa AS (
  SELECT
    s.*,
    json_agg(app.*)->0 application
  FROM
    session s
    JOIN application app ON app.internal_id = s.application_id
  WHERE
    s.is_active IS TRUE
  GROUP BY
    s.internal_id
)
SELECT
  a.internal_id,
  a.id,
  a.email,
  COALESCE(a.phone, ''),
  a.password,
  COALESCE(a.username, ''),
  COALESCE(a.document, ''),
  a.extra_data,
  a.is_active,
  COALESCE(a.pending_email, ''),
  COALESCE(a.pending_phone, ''),
  a.has_email_been_verified,
  a.has_phone_been_verified,
  a.codes,
  json_agg(la.*) links,
  CASE 
    WHEN json_agg(sa.*)::text <> '[null]' 
      THEN json_agg(sa.*)
    ELSE NULL
  END AS active_sessions,
  a.password_updated_at,
  a.created_at,
  a.updated_at
FROM
  user a
  LEFT JOIN la ON la.user_id = a.internal_id
  LEFT JOIN sa ON sa.user_id = a.internal_id
WHERE
  a.id = $1
GROUP BY
  a.internal_id;

-- name: GetUserByEntry :one
WITH la AS (
  SELECT
    l.*,
    json_agg(app.*)->0 application
  FROM
    link l
    JOIN application app ON app.internal_id = l.application_id
  GROUP BY
    l.internal_id
), sa AS (
  SELECT
    s.*,
    json_agg(app.*)->0 application
  FROM
    session s
    JOIN application app ON app.internal_id = s.application_id
  WHERE
    s.is_active IS TRUE
  GROUP BY
    s.internal_id
)
SELECT
  a.internal_id,
  a.id,
  a.email,
  COALESCE(a.phone, ''),
  a.password,
  COALESCE(a.username, ''),
  COALESCE(a.document, ''),
  a.extra_data,
  a.is_active,
  COALESCE(a.pending_email, ''),
  COALESCE(a.pending_phone, ''),
  a.has_email_been_verified,
  a.has_phone_been_verified,
  a.codes,
  json_agg(la.*) links,
  CASE 
    WHEN json_agg(sa.*)::text <> '[null]' 
      THEN json_agg(sa.*)
    ELSE NULL
  END AS active_sessions,
  a.password_updated_at,
  a.created_at,
  a.updated_at
FROM
  user a
  LEFT JOIN la ON la.user_id = a.internal_id
  LEFT JOIN sa ON sa.user_id = a.internal_id 
WHERE
  a.email = $1 OR
  a.phone = $1 OR
  a.username = $1 OR
  a.document = $1
GROUP BY
  a.internal_id;

-- name: GetUserByLink :one
WITH la AS (
  SELECT
    l.*,
    json_agg(app.*)->0 application
  FROM
    link l
    JOIN application app ON app.internal_id = l.application_id
  WHERE
    l.id = $1
  GROUP BY
    l.internal_id
), sa AS (
  SELECT
    s.*,
    json_agg(app.*)->0 application
  FROM
    session s
    JOIN application app ON app.internal_id = s.application_id
  WHERE
    s.is_active IS TRUE
  GROUP BY
    s.internal_id
)
SELECT
  a.internal_id,
  a.id,
  a.email,
  COALESCE(a.phone, ''),
  a.password,
  COALESCE(a.username, ''),
  COALESCE(a.document, ''),
  a.extra_data,
  a.is_active,
  COALESCE(a.pending_email, ''),
  COALESCE(a.pending_phone, ''),
  a.has_email_been_verified,
  a.has_phone_been_verified,
  a.codes,
  json_agg(la.*) links,
  CASE 
    WHEN json_agg(sa.*)::text <> '[null]' 
      THEN json_agg(sa.*)
    ELSE NULL
  END AS active_sessions,
  a.password_updated_at,
  a.created_at,
  a.updated_at
FROM
  user a
  RIGHT JOIN la ON la.user_id = a.internal_id
  LEFT JOIN sa ON sa.user_id = a.internal_id
GROUP BY
  a.internal_id;