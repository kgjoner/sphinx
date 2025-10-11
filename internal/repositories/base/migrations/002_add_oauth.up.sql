ALTER TABLE application
  ADD COLUMN IF NOT EXISTS brand jsonb,
  ADD COLUMN IF NOT EXISTS secret text NOT NULL DEFAULT '$2y$10$9VqcAdmx94/SeDh5ykqsmO0WFKdKXI1nEn95nLiaWRzQNEYZTVm4q',
  ADD COLUMN IF NOT EXISTS allowed_redirect_uris text[];

-- Drop default for secret column if it exists
DO $$
BEGIN
    IF EXISTS (
        SELECT FROM information_schema.columns 
        WHERE table_name = 'application' 
        AND column_name = 'secret' 
        AND column_default IS NOT NULL
        AND table_schema = 'public'
    ) THEN
        ALTER TABLE application ALTER COLUMN secret DROP DEFAULT;
    END IF;
END $$;

ALTER TABLE link
  ADD COLUMN IF NOT EXISTS oauth_code text,
  ADD COLUMN IF NOT EXISTS oauth_expires_at timestamp;