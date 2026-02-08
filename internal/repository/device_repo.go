package repository

import (
	"context"
	"fmt"

	"github.com/ElissaDesign/vpn-dev/internal/errors"
	"github.com/ElissaDesign/vpn-dev/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DeviceRepository handles database operations for devices
type DeviceRepository struct {
	pool *pgxpool.Pool
}

// NewDeviceRepository creates a new device repository
func NewDeviceRepository(pool *pgxpool.Pool) *DeviceRepository {
	return &DeviceRepository{pool: pool}
}

// Create creates a new device in the database
func (r *DeviceRepository) Create(ctx context.Context, device *model.Device) error {
	query := `
		INSERT INTO devices (id, user_id, device_name, device_type, public_key, private_key, ip, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.pool.Exec(ctx, query,
		device.ID,
		device.UserID,
		device.DeviceName,
		device.DeviceType,
		device.PublicKey,
		device.PrivateKey,
		device.IP,
		device.Status,
		device.CreatedAt,
		device.UpdatedAt,
	)

	if err == nil {
		return nil
	}

	// Check for unique constraint violations
	pgErr, ok := err.(*pgconn.PgError)
	if ok {
		if pgErr.ConstraintName == "devices_public_key_key" {
			return errors.ErrDuplicatePublicKey
		}
		if pgErr.ConstraintName == "devices_ip_key" {
			return errors.ErrDuplicateIP
		}
	}

	return fmt.Errorf("%w: %v", errors.ErrDatabase, err)
}

// FindByID retrieves a device by ID
func (r *DeviceRepository) FindByID(ctx context.Context, deviceID uuid.UUID) (*model.Device, error) {
	query := `
		SELECT id, user_id, device_name, device_type, public_key, private_key, host(ip), status, last_connected, created_at, updated_at
		FROM devices
		WHERE id = $1
	`

	var device model.Device
	err := r.pool.QueryRow(ctx, query, deviceID).Scan(
		&device.ID,
		&device.UserID,
		&device.DeviceName,
		&device.DeviceType,
		&device.PublicKey,
		&device.PrivateKey,
		&device.IP,
		&device.Status,
		&device.LastConnected,
		&device.CreatedAt,
		&device.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	return &device, nil
}

// FindByPublicKey retrieves a device by public key
func (r *DeviceRepository) FindByPublicKey(ctx context.Context, publicKey string) (*model.Device, error) {
	query := `
		SELECT id, user_id, device_name, device_type, public_key, private_key, host(ip), status, last_connected, created_at, updated_at
		FROM devices
		WHERE public_key = $1
	`

	var device model.Device
	err := r.pool.QueryRow(ctx, query, publicKey).Scan(
		&device.ID,
		&device.UserID,
		&device.DeviceName,
		&device.DeviceType,
		&device.PublicKey,
		&device.PrivateKey,
		&device.IP,
		&device.Status,
		&device.LastConnected,
		&device.CreatedAt,
		&device.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	return &device, nil
}

// FindByUserID retrieves all devices for a user
func (r *DeviceRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.Device, error) {
	query := `
		SELECT id, user_id, device_name, device_type, public_key, private_key, host(ip), status, last_connected, created_at, updated_at
		FROM devices
		WHERE user_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}
	defer rows.Close()

	var devices []model.Device
	for rows.Next() {
		var device model.Device
		if err := rows.Scan(
			&device.ID,
			&device.UserID,
			&device.DeviceName,
			&device.DeviceType,
			&device.PublicKey,
			&device.PrivateKey,
			&device.IP,
			&device.Status,
			&device.LastConnected,
			&device.CreatedAt,
			&device.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
		}
		devices = append(devices, device)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	return devices, nil
}

// Delete deletes a device
func (r *DeviceRepository) Delete(ctx context.Context, deviceID uuid.UUID) error {
	query := `
		DELETE FROM devices
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, deviceID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}

	return nil
}

// UpdateStatus updates a device's status
func (r *DeviceRepository) UpdateStatus(ctx context.Context, deviceID uuid.UUID, status string) error {
	query := `
		UPDATE devices
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.pool.Exec(ctx, query, status, deviceID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}

	return nil
}

// UpdateLastConnected updates a device's last connected timestamp
func (r *DeviceRepository) UpdateLastConnected(ctx context.Context, deviceID uuid.UUID) error {
	query := `
		UPDATE devices
		SET last_connected = NOW(), updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, deviceID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}

	return nil
}

// CountByUserID counts devices for a user
func (r *DeviceRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) FROM devices WHERE user_id = $1
	`

	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	return count, nil
}
