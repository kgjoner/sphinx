ALTER TABLE application
  ADD COLUMN IF NOT EXISTS brand jsonb;
  ADD COLUMN IF NOT EXISTS secret text NOT NULL DEFAULT '$2y$10$9VqcAdmx94/SeDh5ykqsmO0WFKdKXI1nEn95nLiaWRzQNEYZTVm4q',
  ADD COLUMN IF NOT EXISTS allowed_redirect_uris text[];

ALTER TABLE application
  ALTER COLUMN secret DROP DEFAULT;

ALTER TABLE link
  ADD COLUMN IF NOT EXISTS oauth_code text,
  ADD COLUMN IF NOT EXISTS oauth_expires_at timestamp;