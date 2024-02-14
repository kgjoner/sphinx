-- name: CreateAccount :one
INSERT INTO
  account (
    id,
    email,
    password,
    phone,
    username,
    document,
    is_active,
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
    $10
  )
RETURNING internal_id;

-- name: UpdateAccount :exec
UPDATE
  account
SET
  email = $2,
  password = $3,
  phone = $4,
  username = $5,
  document = $6,
  is_active = $7,
  has_email_been_verified = $8,
  has_phone_been_verified = $9,
  codes = $10,
  password_updated_at = $11,
  updated_at = $12
WHERE
  id = $1;

-- name: GetAccountById :one
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
  a.*,
  json_agg(la.*) links,
  CASE 
    WHEN json_agg(sa.*)::text <> '[null]' 
      THEN json_agg(sa.*)
    ELSE NULL
  END AS active_sessions
FROM
  account a
  LEFT JOIN la ON la.account_id = a.internal_id
  LEFT JOIN sa ON sa.account_id = a.internal_id
WHERE
  a.id = $1
GROUP BY
  a.internal_id;

-- name: GetAccountByEntry :one
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
  a.*,
  json_agg(la.*) links,
  CASE 
    WHEN json_agg(sa.*)::text <> '[null]' 
      THEN json_agg(sa.*)
    ELSE NULL
  END AS active_sessions
FROM
  account a
  LEFT JOIN la ON la.account_id = a.internal_id
  LEFT JOIN sa ON sa.account_id = a.internal_id 
WHERE
  a.email = $1 OR
  a.phone = $1 OR
  a.username = $1 OR
  a.document = $1
GROUP BY
  a.internal_id;
