ALTER TABLE account
  ADD COLUMN IF NOT EXISTS pending_email text,
  ADD COLUMN IF NOT EXISTS pending_phone text;