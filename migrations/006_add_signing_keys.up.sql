CREATE TABLE IF NOT EXISTS signing_key (
    internal_id INT GENERATED ALWAYS AS IDENTITY,
    id UUID NOT NULL UNIQUE,
    key_id VARCHAR(255) NOT NULL UNIQUE,
    algorithm VARCHAR(10) NOT NULL,
    public_key TEXT NOT NULL,
    private_key TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    activates_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    rotated_at TIMESTAMPTZ,

    PRIMARY KEY (internal_id)
);

CREATE INDEX IF NOT EXISTS idx_signing_key_active ON signing_key(is_active, expires_at) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_signing_key_kid ON signing_key(key_id);

-- Seed initial HS256 compatibility key entry for graceful migration
-- This allows existing HS256 tokens to be validated during the first rotation period
INSERT INTO signing_key (id, key_id, algorithm, public_key, private_key, is_active, expires_at)
VALUES (
    '00000000-0000-0000-0000-000000000000',
    'hs256-legacy',
    'HS256',
    '',  -- HS256 doesn't use public key
    '',  -- Will use JWT.SECRET from env config for validation
    true,
    NOW() + INTERVAL '6 months'  -- Expires after default refresh token lifetime
);
