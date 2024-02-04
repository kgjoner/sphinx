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
SELECT
  a.*,
  json_agg(l.*) links,
  json_agg(s.*) active_sessions
FROM
  account a
  LEFT JOIN session s ON s.account_id = a.id AND s.is_active IS TRUE
  LEFT JOIN link l ON l.account_id = a.id 
WHERE
  a.id = $1
GROUP BY
  a.id;

-- name: GetAccountByEntry :one
SELECT
  a.*,
  json_agg(l.*) links,
  json_agg(s.*) active_sessions
FROM
  account a
  LEFT JOIN session s ON s.account_id = a.internal_id AND s.is_active IS TRUE
  LEFT JOIN link l ON l.account_id = a.internal_id 
WHERE
  a.email = $1 OR
  a.phone = $1 OR
  a.username = $1 OR
  a.document = $1
GROUP BY
  a.internal_id;
