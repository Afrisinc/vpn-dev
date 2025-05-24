package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Subscription struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type      string             `bson:"type" json:"type"`     // e.g., "monthly"
	Status    string             `bson:"status" json:"status"` // e.g., "active", "expired"
	ExpiresAt time.Time          `bson:"expiresAt" json:"expiresAt"`
}
