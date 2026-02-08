-- Migrate existing users to device model
INSERT INTO devices (user_id, device_name, device_type, public_key, private_key, ip, status, last_connected, created_at, updated_at)
SELECT id, 'Primary Device', 'unknown', public_key, private_key, ip, status, last_connected, created_at, updated_at
FROM users
WHERE public_key IS NOT NULL AND ip IS NOT NULL AND private_key IS NOT NULL;
