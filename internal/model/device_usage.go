package model

import (
	"time"

	"github.com/google/uuid"
)

// DeviceUsage represents bandwidth usage snapshot for a device
type DeviceUsage struct {
	ID             uuid.UUID `db:"id" json:"id"`
	DeviceID       uuid.UUID `db:"device_id" json:"deviceId"`
	UserID         uuid.UUID `db:"user_id" json:"userId"`
	BytesSent      int64     `db:"bytes_sent" json:"bytesSent"`
	BytesReceived  int64     `db:"bytes_received" json:"bytesReceived"`
	RecordedAt     time.Time `db:"recorded_at" json:"recordedAt"`
}

// DeviceUsageSummary represents total usage for a device
type DeviceUsageSummary struct {
	DeviceID      uuid.UUID `json:"deviceId"`
	DeviceName    string    `json:"deviceName"`
	BytesSent     int64     `json:"bytesSent"`
	BytesReceived int64     `json:"bytesReceived"`
	TotalBytes    int64     `json:"totalBytes"`
	TotalGB       float64   `json:"totalGB"`
}

// UserUsageSummary represents total usage for all user devices
type UserUsageSummary struct {
	TotalBytes    int64                   `json:"totalBytes"`
	TotalGB       float64                 `json:"totalGB"`
	BytesSent     int64                   `json:"bytesSent"`
	BytesReceived int64                   `json:"bytesReceived"`
	DeviceUsages  []DeviceUsageSummary    `json:"deviceUsages,omitempty"`
}
