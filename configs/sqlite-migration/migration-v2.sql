ALTER TABLE luna4_users
DROP COLUMN last_login_at;

ALTER TABLE luna4_email_auth
DROP COLUMN expires_at;
