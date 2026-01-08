-- -----------------------------
-- Down Migration: Drop Auth & Users Tables
-- -----------------------------

-- Drop auth_credentials first (FK depends on users)
DROP TABLE IF EXISTS auth_credentials;

DROP TABLE IF EXISTS users;
