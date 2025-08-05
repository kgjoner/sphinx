ALTER TABLE account
  RENAME TO user
  ADD column IF NOT EXISTS external_auth_ids jsonb;

ALTER TABLE application
  RENAME COLUMN IF EXISTS grantings TO possible_roles,
  DROP COLUMN IF EXISTS brand;

UPDATE link
  SET has_consent = true,
  SET roles = CONCAT(roles, grantings);

ALTER TABLE link
  ADD COLUMN IF NOT EXISTS has_consent boolean NOT NULL DEFAULT false,
  DROP COLUMN IF EXISTS grantings,
  DROP COLUMN IF EXISTS oauth_code,
  DROP COLUMN IF EXISTS oauth_expires_at,
  RENAME COLUMN IF EXISTS account_id TO user_id,
  ADD CONSTRAINT app_user_unique UNIQUE (application_id, user_id);

ALTER TABLE session
  RENAME COLUMN IF EXISTS account_id TO user_id;