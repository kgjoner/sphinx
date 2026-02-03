-- name: InsertSigningKey :exec
INSERT INTO
    signing_key (
        id,
        key_id,
        algorithm,
        public_key,
        private_key,
        is_active,
        activates_at,
        expires_at,
        rotated_at
    )
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: UpdateSigningKey :exec
UPDATE signing_key
SET
    is_active = $1,
    expires_at = $2,
    rotated_at = $3
WHERE
    id = $1;

-- name: ListActiveSigningKeys :many
SELECT
    id,
    key_id,
    algorithm,
    public_key,
    private_key,
    is_active,
    created_at,
    activates_at,
    expires_at,
    rotated_at
FROM
    signing_key
WHERE
    is_active = true
    AND (
        expires_at IS NULL
        OR expires_at > NOW()
    )
ORDER BY
    created_at DESC;

-- name: GetSigningKeyByKID :one
SELECT
    id,
    key_id,
    algorithm,
    public_key,
    private_key,
    is_active,
    created_at,
    activates_at,
    expires_at,
    rotated_at
FROM
    signing_key
WHERE
    key_id = $1
LIMIT 1;

-- name: DeactivateExpiredKeys :exec
UPDATE signing_key
SET
    is_active = false
WHERE
    expires_at IS NOT NULL
    AND expires_at < NOW()
    AND is_active = true;
