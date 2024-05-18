ALTER TABLE application
  DROP COLUMN IF EXISTS secret,
  DROP COLUMN IF EXISTS allowed_redirect_uris;