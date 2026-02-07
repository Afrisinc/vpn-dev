-- Remove server_id from users table
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_server_id;
DROP INDEX IF EXISTS idx_users_server_id;
ALTER TABLE users DROP COLUMN IF EXISTS server_id;
