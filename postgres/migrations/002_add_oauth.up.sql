ALTER TABLE application
  ADD COLUMN IF NOT EXISTS secret text NOT NULL,
  ADD COLUMN IF NOT EXISTS allowed_redirect_uris text[];