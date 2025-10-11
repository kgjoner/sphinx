CREATE TABLE IF NOT EXISTS account (
  internal_id int GENERATED ALWAYS AS IDENTITY,  
  id uuid NOT NULL UNIQUE,
  email text NOT NULL UNIQUE,
  password text NOT NULL,
  phone text UNIQUE,
  username text UNIQUE,
  document text UNIQUE,
  extra_data jsonb,

  is_active bool NOT NULL,
  has_email_been_verified bool NOT NULL,
  has_phone_been_verified bool NOT NULL,
  codes jsonb NOT NULL,

  password_updated_at timestamp,
  created_at timestamp NOT NULL DEFAULT NOW(),
  updated_at timestamp NOT NULL DEFAULT NOW(),

  PRIMARY KEY(internal_id)
);

CREATE TABLE IF NOT EXISTS application (
  internal_id int GENERATED ALWAYS AS IDENTITY,  
  id uuid NOT NULL UNIQUE,
  name text NOT NULL UNIQUE,
  grantings text[] NOT NULL,

  created_at timestamp NOT NULL DEFAULT NOW(),
  updated_at timestamp NOT NULL DEFAULT NOW(),

  PRIMARY KEY(internal_id)
);

CREATE TABLE IF NOT EXISTS link (
  internal_id int GENERATED ALWAYS AS IDENTITY,  
  id uuid NOT NULL UNIQUE,
  account_id int NOT NULL,
  application_id int NOT NULL,
  roles text[],
  grantings text[],

  created_at timestamp NOT NULL DEFAULT NOW(),
  updated_at timestamp NOT NULL DEFAULT NOW(),

  PRIMARY KEY(internal_id),
  CONSTRAINT fk_account
    FOREIGN KEY(account_id)
      REFERENCES account(internal_id)
      ON DELETE CASCADE,
  CONSTRAINT fk_application
    FOREIGN KEY(application_id)
      REFERENCES application(internal_id)
      ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS session (
  internal_id int GENERATED ALWAYS AS IDENTITY,  
  id uuid NOT NULL UNIQUE,
  account_id int NOT NULL,
  application_id int NOT NULL,
  refresh_token text,
  refreshed_at timestamp,
  elapsed_minutes_between_refreshes int[] NOT NULL,
  refreshes_count int NOT NULL,
  device text NOT NULL,
  ip text NOT NULL,

  is_active bool,
  terminated_at timestamp,
  created_at timestamp NOT NULL DEFAULT NOW(),
  updated_at timestamp NOT NULL DEFAULT NOW(),

  PRIMARY KEY(internal_id),
  CONSTRAINT fk_account
    FOREIGN KEY(account_id)
      REFERENCES account(internal_id)
      ON DELETE CASCADE,
  CONSTRAINT fk_application
    FOREIGN KEY(application_id)
      REFERENCES application(internal_id)
      ON DELETE CASCADE
);

INSERT INTO 
  application(
    id,
    name,
    grantings
  )
VALUES (
  '80cadd74-5ccd-41c4-9938-3c8961be04db',
  'sphynx',
  '{"ADMIN","DEV"}'
)
ON CONFLICT (id)
DO NOTHING;