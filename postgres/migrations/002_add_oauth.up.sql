ALTER TABLE application
  ADD COLUMN IF NOT EXISTS secret text NOT NULL,
  ADD COLUMN IF NOT EXISTS allowed_redirect_uris text[];

ALTER TABLE link
  ADD COLUMN IF NOT EXISTS oauth_code text,
  ADD COLUMN IF NOT EXISTS oauth_expires_at timestamp;