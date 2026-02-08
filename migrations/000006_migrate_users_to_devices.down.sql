-- Remove migrated devices
DELETE FROM devices WHERE device_name = 'Primary Device' AND device_type = 'unknown';
