ALTER TABLE link
  ADD COLUMN IF NOT EXISTS oauth_code_challenge text,
  ADD COLUMN IF NOT EXISTS oauth_code_challenge_method text;
