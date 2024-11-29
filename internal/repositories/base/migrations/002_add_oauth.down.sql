ALTER TABLE application
  DROP COLUMN IF EXISTS secret,
  DROP COLUMN IF EXISTS allowed_redirect_uris;

ALTER TABLE link
  DROP COLUMN IF EXISTS oauth_code,
  DROP COLUMN IF EXISTS oauth_expires_at;