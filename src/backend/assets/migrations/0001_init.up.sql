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