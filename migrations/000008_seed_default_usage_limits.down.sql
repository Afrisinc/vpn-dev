-- Revert default data usage limits
-- This is a seed migration, so down just resets to default if needed
-- In practice, you might not need to do anything for a down migration on a seed

-- If you want to reset all limits back to a previous value:
-- UPDATE users SET data_usage_limit = 0 WHERE data_usage_limit = 107374182400;
