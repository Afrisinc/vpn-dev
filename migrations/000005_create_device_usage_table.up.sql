CREATE TABLE IF NOT EXISTS device_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bytes_sent BIGINT DEFAULT 0,
    bytes_received BIGINT DEFAULT 0,
    recorded_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_device_usage_device_id ON device_usage(device_id);
CREATE INDEX idx_device_usage_user_id ON device_usage(user_id);
CREATE INDEX idx_device_usage_recorded_at ON device_usage(recorded_at DESC);
