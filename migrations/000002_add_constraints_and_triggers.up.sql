ALTER TABLE users
    ADD CONSTRAINT users_role_check
        CHECK (role IN ('user', 'admin', 'therapist')),
    ADD CONSTRAINT users_status_check
        CHECK (status IN ('pending', 'active', 'suspended', 'deleted'));

ALTER TABLE auth_credentials
    ADD CONSTRAINT auth_credentials_failed_attempts_non_negative
        CHECK (failed_attempts >= 0);

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER auth_credentials_updated_at
    BEFORE UPDATE ON auth_credentials
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
