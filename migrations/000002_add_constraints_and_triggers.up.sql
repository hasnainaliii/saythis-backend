-- ============================================================
-- CHECK constraints
-- The application already validates these values, but DB-level
-- constraints are a second line of defence: nothing that bypasses
-- the app layer (scripts, direct SQL, future bugs) can insert
-- an invalid role or status.
-- ============================================================

ALTER TABLE users
    ADD CONSTRAINT users_role_check
        CHECK (role IN ('user', 'admin', 'therapist')),
    ADD CONSTRAINT users_status_check
        CHECK (status IN ('active', 'suspended', 'deleted'));

ALTER TABLE auth_credentials
    ADD CONSTRAINT auth_credentials_failed_attempts_non_negative
        CHECK (failed_attempts >= 0);

-- ============================================================
-- updated_at auto-update trigger
-- Without this, updated_at only changes when the application
-- explicitly sets it.  The trigger makes it impossible to forget.
-- ============================================================

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
