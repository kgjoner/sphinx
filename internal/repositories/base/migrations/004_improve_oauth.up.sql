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
  ADD CONSTRAINT app_acc_unique UNIQUE (application_id, account_id);
