-- Add data usage limit column to users table
-- Stores limit in bytes (0 = unlimited)
ALTER TABLE users
ADD COLUMN data_usage_limit BIGINT NOT NULL DEFAULT 107374182400; -- Default: 100 GB in bytes

-- Create index for queries filtering by limit
CREATE INDEX idx_users_data_usage_limit ON users(data_usage_limit);
