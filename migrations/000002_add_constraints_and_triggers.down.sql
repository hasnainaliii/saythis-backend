DROP TRIGGER IF EXISTS auth_credentials_updated_at ON auth_credentials;
DROP TRIGGER IF EXISTS users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at;

ALTER TABLE auth_credentials
    DROP CONSTRAINT IF EXISTS auth_credentials_failed_attempts_non_negative;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_status_check,
    DROP CONSTRAINT IF EXISTS users_role_check;
