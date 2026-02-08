-- Create servers table for multi-server VPN infrastructure
CREATE TABLE IF NOT EXISTS servers (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    location VARCHAR(100) NOT NULL,
    region_code VARCHAR(20) NOT NULL,
    country_code VARCHAR(2) NOT NULL,

    -- Connection info
    public_ip VARCHAR(45) NOT NULL,
    agent_url VARCHAR(255) NOT NULL,
    agent_api_key TEXT NOT NULL,
    wireguard_port INT DEFAULT 51820,

    -- WireGuard specific
    server_public_key TEXT NOT NULL,
    network_cidr VARCHAR(20) DEFAULT '192.168.88.0/24',
    next_available_ip INT DEFAULT 2,

    -- Capacity & limits
    max_clients INT DEFAULT 250,
    current_clients INT DEFAULT 0,
    bandwidth_limit_mbps INT,

    -- Status & health
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'maintenance', 'offline', 'full')),
    last_health_check TIMESTAMP,
    health_status VARCHAR(20) DEFAULT 'unknown' CHECK (health_status IN ('healthy', 'degraded', 'down', 'unknown')),

    -- Geo info
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_servers_status ON servers(status);
CREATE INDEX idx_servers_region ON servers(region_code);
CREATE INDEX idx_servers_country ON servers(country_code);
CREATE INDEX idx_servers_health ON servers(health_status);
