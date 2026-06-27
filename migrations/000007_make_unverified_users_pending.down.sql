UPDATE users
SET status = 'active'
WHERE status = 'pending';

ALTER TABLE users
    ALTER COLUMN status SET DEFAULT 'active';

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_status_check,
    ADD CONSTRAINT users_status_check
        CHECK (status IN ('active', 'suspended', 'deleted'));
