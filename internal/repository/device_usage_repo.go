package repository

import (
	"context"
	"fmt"

	"github.com/ElissaDesign/vpn-dev/internal/errors"
	"github.com/ElissaDesign/vpn-dev/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DeviceUsageRepository handles database operations for device usage
type DeviceUsageRepository struct {
	pool *pgxpool.Pool
}

// NewDeviceUsageRepository creates a new device usage repository
func NewDeviceUsageRepository(pool *pgxpool.Pool) *DeviceUsageRepository {
	return &DeviceUsageRepository{pool: pool}
}

// RecordUsage records a usage snapshot for a device
func (r *DeviceUsageRepository) RecordUsage(ctx context.Context, usage *model.DeviceUsage) error {
	query := `
		INSERT INTO device_usage (id, device_id, user_id, bytes_sent, bytes_received, recorded_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	err := r.pool.QueryRow(ctx, query,
		usage.ID,
		usage.DeviceID,
		usage.UserID,
		usage.BytesSent,
		usage.BytesReceived,
		usage.RecordedAt,
	).Scan()

	if err == nil {
		return nil
	}

	return fmt.Errorf("%w: %v", errors.ErrDatabase, err)
}

// GetLatestUsage gets the latest usage record for a device
func (r *DeviceUsageRepository) GetLatestUsage(ctx context.Context, deviceID uuid.UUID) (*model.DeviceUsage, error) {
	query := `
		SELECT id, device_id, user_id, bytes_sent, bytes_received, recorded_at
		FROM device_usage
		WHERE device_id = $1
		ORDER BY recorded_at DESC
		LIMIT 1
	`

	var usage model.DeviceUsage
	err := r.pool.QueryRow(ctx, query, deviceID).Scan(
		&usage.ID,
		&usage.DeviceID,
		&usage.UserID,
		&usage.BytesSent,
		&usage.BytesReceived,
		&usage.RecordedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil // No usage data yet is not an error
		}
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	return &usage, nil
}

// GetTotalUsageByDevice gets total usage for a device
func (r *DeviceUsageRepository) GetTotalUsageByDevice(ctx context.Context, deviceID uuid.UUID) (*model.DeviceUsageSummary, error) {
	query := `
		SELECT
			device_id,
			COALESCE(SUM(bytes_sent), 0) as total_sent,
			COALESCE(SUM(bytes_received), 0) as total_received
		FROM device_usage
		WHERE device_id = $1
		GROUP BY device_id
	`

	var summary model.DeviceUsageSummary
	var deviceID_result uuid.UUID

	err := r.pool.QueryRow(ctx, query, deviceID).Scan(
		&deviceID_result,
		&summary.BytesSent,
		&summary.BytesReceived,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			// No usage data yet, return empty summary
			summary.DeviceID = deviceID
			summary.BytesSent = 0
			summary.BytesReceived = 0
			summary.TotalBytes = 0
			summary.TotalGB = 0
			return &summary, nil
		}
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	summary.DeviceID = deviceID_result
	summary.TotalBytes = summary.BytesSent + summary.BytesReceived
	summary.TotalGB = float64(summary.TotalBytes) / (1024 * 1024 * 1024)

	return &summary, nil
}

// GetTotalUsageByUser gets total usage for all user devices
func (r *DeviceUsageRepository) GetTotalUsageByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `
		SELECT COALESCE(SUM(bytes_sent + bytes_received), 0) as total
		FROM device_usage
		WHERE user_id = $1
	`

	var total int64
	err := r.pool.QueryRow(ctx, query, userID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	return total, nil
}

// GetUserUsageSummary gets detailed usage summary for all user devices
func (r *DeviceUsageRepository) GetUserUsageSummary(ctx context.Context, userID uuid.UUID) (*model.UserUsageSummary, error) {
	query := `
		SELECT
			d.id,
			d.device_name,
			COALESCE(SUM(du.bytes_sent), 0) as total_sent,
			COALESCE(SUM(du.bytes_received), 0) as total_received
		FROM devices d
		LEFT JOIN device_usage du ON du.device_id = d.id
		WHERE d.user_id = $1
		GROUP BY d.id, d.device_name
		ORDER BY d.created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}
	defer rows.Close()

	summary := &model.UserUsageSummary{
		DeviceUsages: []model.DeviceUsageSummary{},
	}

	for rows.Next() {
		var deviceID uuid.UUID
		var deviceName string
		var bytesSent, bytesReceived int64

		if err := rows.Scan(&deviceID, &deviceName, &bytesSent, &bytesReceived); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
		}

		deviceUsage := model.DeviceUsageSummary{
			DeviceID:      deviceID,
			DeviceName:    deviceName,
			BytesSent:     bytesSent,
			BytesReceived: bytesReceived,
			TotalBytes:    bytesSent + bytesReceived,
			TotalGB:       float64(bytesSent+bytesReceived) / (1024 * 1024 * 1024),
		}

		summary.DeviceUsages = append(summary.DeviceUsages, deviceUsage)
		summary.BytesSent += bytesSent
		summary.BytesReceived += bytesReceived
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	summary.TotalBytes = summary.BytesSent + summary.BytesReceived
	summary.TotalGB = float64(summary.TotalBytes) / (1024 * 1024 * 1024)

	return summary, nil
}
