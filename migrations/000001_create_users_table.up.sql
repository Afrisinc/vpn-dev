-- Create users table with WireGuard configuration
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    ip INET NOT NULL UNIQUE,
    private_key TEXT NOT NULL,
    public_key TEXT NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'disconnected',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_connected TIMESTAMP,

    CONSTRAINT valid_status CHECK (status IN ('active', 'disconnected'))
);

-- Create indexes for common queries
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_public_key ON users(public_key);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_ip ON users(ip);
CREATE INDEX idx_users_created_at ON users(created_at);
