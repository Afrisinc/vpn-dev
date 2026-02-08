-- Seed default data usage limits for existing users
-- Default limits: 100 GB per user
-- This migration runs after adding the column, so it only applies to existing records

-- Ensure all users have the default 100 GB (107374182400 bytes) limit if not already set
UPDATE users
SET data_usage_limit = 107374182400
WHERE data_usage_limit IS NULL OR data_usage_limit = 0;
