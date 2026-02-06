-- 1. Enable required extensions
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE EXTENSION IF NOT EXISTS unaccent;

CREATE OR REPLACE FUNCTION my_unaccent(text) RETURNS text AS $$
    SELECT unaccent($1);
$$ LANGUAGE sql IMMUTABLE;

-- 2. Add a generated column that combines the searched fields
ALTER TABLE "user"
ADD COLUMN username_updated_at timestamp,
ADD COLUMN search_field text GENERATED ALWAYS AS (
  my_unaccent(
    COALESCE(username, '') || ' ' || 
    COALESCE(extra_data->>'name', '') || ' ' || 
    COALESCE(extra_data->>'surname', '') || ' ' || 
    split_part(email, '@', 1)
  )
) STORED;

-- 3. Create an index for performance
CREATE INDEX idx_user_search_tgrm ON "user" USING gin (search_field gin_trgm_ops);