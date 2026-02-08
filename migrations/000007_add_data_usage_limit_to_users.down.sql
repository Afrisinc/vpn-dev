-- Remove data usage limit column from users table
DROP INDEX IF EXISTS idx_users_data_usage_limit;
ALTER TABLE users DROP COLUMN data_usage_limit;
