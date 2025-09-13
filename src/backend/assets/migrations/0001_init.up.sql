CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,
    user_name TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    hashed_password TEXT NOT NULL UNIQUE,
    hashed_cookie_value TEXT UNIQUE,
    expiration_date TEXT,
    used_space_in_bytes BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS apps (
    app_id SERIAL PRIMARY KEY,
    user_id INTEGER,
    app_name TEXT,
    UNIQUE(user_id, app_name),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS versions (
    version_id SERIAL PRIMARY KEY,
    version_name TEXT NOT NULL,
    app_id INTEGER NOT NULL,
    creation_timestamp TIMESTAMPTZ NOT NULL,
    data BYTEA NOT NULL,
    UNIQUE(app_id, version_id),
    FOREIGN KEY (app_id) REFERENCES apps(app_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS configs (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Enable trigram support for fast %...% pattern matching
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Trigram GIN indexes speed up ILIKE/LIKE '%term%' on names
CREATE INDEX IF NOT EXISTS users_user_name_trgm ON users USING gin (user_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS apps_app_name_trgm   ON apps  USING gin (app_name  gin_trgm_ops);

-- Supports JOIN apps a ON a.user_id = u.user_id
CREATE INDEX IF NOT EXISTS apps_user_id_idx ON apps(user_id);

-- Lets Postgres get the newest version per app via index order
CREATE INDEX IF NOT EXISTS versions_app_id_created_desc ON versions (app_id, creation_timestamp DESC);

-- Optional: supports LOWER(col) predicates if used; trigram remains best for %...%
CREATE INDEX IF NOT EXISTS users_user_name_lower_idx ON users (LOWER(user_name));
CREATE INDEX IF NOT EXISTS apps_app_name_lower_idx   ON apps  (LOWER(app_name));
