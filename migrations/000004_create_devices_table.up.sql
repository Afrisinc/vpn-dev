CREATE TABLE IF NOT EXISTS devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_name VARCHAR(100) NOT NULL,
    device_type VARCHAR(50) DEFAULT 'unknown',
    public_key TEXT NOT NULL UNIQUE,
    private_key TEXT NOT NULL,
    ip INET NOT NULL UNIQUE,
    status VARCHAR(20) DEFAULT 'disconnected',
    last_connected TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT check_device_status CHECK (status IN ('active', 'disconnected', 'suspended')),
    CONSTRAINT check_device_type CHECK (device_type IN ('mobile', 'desktop', 'tablet', 'router', 'unknown'))
);

CREATE INDEX idx_devices_user_id ON devices(user_id);
CREATE INDEX idx_devices_public_key ON devices(public_key);
CREATE INDEX idx_devices_status ON devices(status);
CREATE INDEX idx_devices_created_at ON devices(created_at);
