ALTER TABLE session
  RENAME COLUMN IF EXISTS user_id TO account_id;

ALTER TABLE link
  DROP CONSTRAINT IF EXISTS app_user_unique,
  RENAME COLUMN IF EXISTS user_id TO account_id,
  ADD COLUMN IF NOT EXISTS grantings text[] NOT NULL,
  ADD COLUMN IF NOT EXISTS oauth_code text,
  ADD COLUMN IF NOT EXISTS oauth_expires_at timestamp,
  DROP COLUMN IF EXISTS has_consent;

-- Note: This data migration is lossy - we cannot fully restore the original grantings
-- if they were concatenated with roles during the up migration
UPDATE link
  SET grantings = roles;

ALTER TABLE application
  ADD COLUMN IF NOT EXISTS brand jsonb,
  RENAME COLUMN IF EXISTS possible_roles TO grantings;

ALTER TABLE user
  DROP COLUMN IF EXISTS external_auth_ids,
  RENAME TO account;