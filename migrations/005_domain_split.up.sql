CREATE TABLE IF NOT EXISTS external_credential (
    internal_id int GENERATED ALWAYS AS IDENTITY,
    user_id int NOT NULL REFERENCES "user"(internal_id) ON DELETE CASCADE,
    provider_name VARCHAR(255) NOT NULL,
    provider_subject_id VARCHAR(255) NOT NULL,
    provider_alias VARCHAR(255),
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (internal_id),
    UNIQUE (provider_name, provider_subject_id)
);