-- Add index on auth_credentials.user_id for faster lookups
CREATE INDEX idx_auth_credentials_user_id ON auth_credentials(user_id);
