ALTER TABLE users
    ALTER COLUMN status SET DEFAULT 'pending';

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_status_check,
    ADD CONSTRAINT users_status_check
        CHECK (status IN ('pending', 'active', 'suspended', 'deleted'));

UPDATE users
SET status = 'pending'
WHERE status = 'active'
  AND email_verified_at IS NULL;
