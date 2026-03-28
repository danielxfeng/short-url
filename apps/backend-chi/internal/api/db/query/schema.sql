CREATE TYPE provider_enum AS ENUM ('GOOGLE', 'GITHUB');

CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  provider provider_enum NOT NULL,
  provider_id TEXT NOT NULL,
  display_name TEXT,
  profile_pic TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (provider, provider_id)
);

CREATE TABLE links (
  id SERIAL PRIMARY KEY,
  user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  code TEXT NOT NULL UNIQUE,
  original_url TEXT NOT NULL,
  clicks INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
)
