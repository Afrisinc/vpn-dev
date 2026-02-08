-- Add server_id to users table for multi-server support
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'server_id'
    ) THEN
        ALTER TABLE users ADD COLUMN server_id VARCHAR(50);
    END IF;
END $$;

-- Create foreign key constraint if it doesn't exist
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'fk_users_server_id'
    ) THEN
        ALTER TABLE users
        ADD CONSTRAINT fk_users_server_id
        FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE RESTRICT;
    END IF;
END $$;

-- Create index if it doesn't exist
CREATE INDEX IF NOT EXISTS idx_users_server_id ON users(server_id);
