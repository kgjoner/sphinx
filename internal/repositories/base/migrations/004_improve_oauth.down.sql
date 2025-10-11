-- Rename user_id back to account_id in session table if it exists
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'session' AND column_name = 'user_id' AND table_schema = 'public'
    ) THEN
        ALTER TABLE session RENAME COLUMN user_id TO account_id;
    END IF;
END $$;

ALTER TABLE link
  DROP CONSTRAINT IF EXISTS link_application_user_unique,
  DROP CONSTRAINT IF EXISTS app_user_unique,
  ADD COLUMN IF NOT EXISTS grantings text[],
  ADD COLUMN IF NOT EXISTS oauth_code text,
  ADD COLUMN IF NOT EXISTS oauth_expires_at timestamp,
  DROP COLUMN IF EXISTS has_consent;

-- Rename user_id back to account_id in link table if it exists
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'link' AND column_name = 'user_id' AND table_schema = 'public'
    ) THEN
        ALTER TABLE link RENAME COLUMN user_id TO account_id;
    END IF;
END $$;

-- Note: This data migration is lossy - we cannot fully restore the original grantings
-- if they were concatenated with roles during the up migration
UPDATE link
  SET grantings = roles;

ALTER TABLE application
  ADD COLUMN IF NOT EXISTS brand jsonb;

-- Rename possible_roles back to grantings in application table if it exists
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'application' AND column_name = 'possible_roles' AND table_schema = 'public'
    ) THEN
        ALTER TABLE application RENAME COLUMN possible_roles TO grantings;
    END IF;
END $$;

-- Drop external_auth_ids column from user or account table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'user' AND table_schema = 'public'
    ) THEN
        ALTER TABLE "user" DROP COLUMN IF EXISTS external_auth_ids;
    ELSIF EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'account' AND table_schema = 'public'
    ) THEN
        ALTER TABLE account DROP COLUMN IF EXISTS external_auth_ids;
    END IF;
END $$;

-- Rename user table back to account if it exists
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'user' AND table_schema = 'public'
    ) THEN
        ALTER TABLE "user" RENAME TO account;
    END IF;
END $$;