ALTER TABLE link
  DROP COLUMN IF EXISTS oauth_code_challenge,
  DROP COLUMN IF EXISTS oauth_code_challenge_method;
