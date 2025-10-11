-- Rename account table to user if it hasn't been renamed yet
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'account' AND table_schema = 'public'
    ) THEN
        ALTER TABLE account RENAME TO "user";
    END IF;
END $$;

-- Add external_auth_ids column to user table
ALTER TABLE "user"
  ADD COLUMN IF NOT EXISTS external_auth_ids jsonb;

ALTER TABLE application
  DROP COLUMN IF EXISTS brand;

-- Rename grantings to possible_roles in application table if it exists
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'application' AND column_name = 'grantings' AND table_schema = 'public'
    ) THEN
        ALTER TABLE application RENAME COLUMN grantings TO possible_roles;
    END IF;
END $$;

-- Update link table data before dropping columns
-- Only update if grantings column exists and has_consent doesn't exist yet
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'link' AND column_name = 'grantings' AND table_schema = 'public'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'link' AND column_name = 'has_consent' AND table_schema = 'public'
    ) THEN
        UPDATE link
        SET roles = COALESCE(roles, '{}') || COALESCE(grantings, '{}');
    END IF;
END $$;

ALTER TABLE link
  ADD COLUMN IF NOT EXISTS has_consent boolean NOT NULL DEFAULT false,
  DROP COLUMN IF EXISTS grantings,
  DROP COLUMN IF EXISTS oauth_code,
  DROP COLUMN IF EXISTS oauth_expires_at;

-- Rename account_id to user_id in link table if it hasn't been renamed yet
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'link' AND column_name = 'account_id' AND table_schema = 'public'
    ) THEN
        ALTER TABLE link RENAME COLUMN account_id TO user_id;
    END IF;
END $$;

-- Add unique constraint for application_id and user_id if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE table_name = 'link' 
        AND constraint_name = 'link_application_user_unique'
        AND table_schema = 'public'
    ) THEN
        ALTER TABLE link ADD CONSTRAINT link_application_user_unique UNIQUE (application_id, user_id);
    END IF;
END $$;

-- Rename account_id to user_id in session table if it hasn't been renamed yet
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'session' AND column_name = 'account_id' AND table_schema = 'public'
    ) THEN
        ALTER TABLE session RENAME COLUMN account_id TO user_id;
    END IF;
END $$;