-- Luna4User table
CREATE TABLE IF NOT EXISTS luna4_users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    status TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    last_login_at INTEGER
);

-- Luna4EmailAuth table
CREATE TABLE IF NOT EXISTS luna4_email_auth (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token TEXT NOT NULL,
    sent_at INTEGER NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES luna4_users(id) ON DELETE CASCADE
);

-- Luna4UserService table
CREATE TABLE IF NOT EXISTS luna4_user_service (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    service TEXT NOT NULL,
    permission TEXT NOT NULL,
    expires_at INTEGER,
    FOREIGN KEY (user_id) REFERENCES luna4_users(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_luna4_users_email ON luna4_users(email);
CREATE INDEX IF NOT EXISTS idx_luna4_users_status ON luna4_users(status);
CREATE INDEX IF NOT EXISTS idx_luna4_email_auth_user_id ON luna4_email_auth(user_id);
CREATE INDEX IF NOT EXISTS idx_luna4_email_auth_token ON luna4_email_auth(token);
CREATE INDEX IF NOT EXISTS idx_luna4_user_service_user_id ON luna4_user_service(user_id);
CREATE INDEX IF NOT EXISTS idx_luna4_user_service_service ON luna4_user_service(service);

PRAGMA schema_version = 1;