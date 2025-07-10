ALTER TABLE account
  DROP COLUMN IF EXISTS pending_email,
  DROP COLUMN IF EXISTS pending_phone;