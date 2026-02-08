package model

import (
	"time"

	"github.com/google/uuid"
)

// Device represents a VPN device belonging to a user
type Device struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	UserID        uuid.UUID  `db:"user_id" json:"userId"`
	DeviceName    string     `db:"device_name" json:"deviceName" validate:"required,min=1,max=100"`
	DeviceType    string     `db:"device_type" json:"deviceType" validate:"required,oneof=mobile desktop tablet router unknown"`
	PublicKey     string     `db:"public_key" json:"publicKey"`
	PrivateKey    string     `db:"private_key" json:"privateKey,omitempty"`
	IP            string     `db:"ip" json:"ip"`
	Status        string     `db:"status" json:"status" validate:"oneof=active disconnected suspended"`
	LastConnected *time.Time `db:"last_connected" json:"lastConnected,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updatedAt"`
}
