package model

import (
	"database/sql/driver"
	"time"

	"github.com/google/uuid"
)

// User represents a VPN user with WireGuard configuration
type User struct {
	ID                uuid.UUID  `db:"id" json:"id" validate:"required,uuid"`
	Email             string     `db:"email" json:"email" validate:"required,email"`
	IP                string     `db:"ip" json:"ip" validate:"required,ipv4"`
	PrivateKey        string     `db:"private_key" json:"privateKey" validate:"required"`
	PublicKey         string     `db:"public_key" json:"publicKey" validate:"required"`
	Status            string     `db:"status" json:"status" validate:"oneof=active disconnected"`
	ServerID          *string    `db:"server_id" json:"serverId,omitempty" validate:"omitempty"`
	DataUsageLimit    int64      `db:"data_usage_limit" json:"dataUsageLimit" validate:"required,min=0"`
	CreatedAt         time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updatedAt"`
	LastConnected     *time.Time `db:"last_connected" json:"lastConnected,omitempty"`
}

// Scan implements sql.Scanner interface for scanning database values
func (u *User) Scan(value interface{}) error {
	return nil // Implement if using row.Scan directly
}

// Value implements driver.Valuer interface for writing to database
func (u User) Value() (driver.Value, error) {
	return u.ID, nil // Implement if needed
}
