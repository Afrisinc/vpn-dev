package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email        string             `bson:"email" json:"email"`
	IP           string             `bson:"ip" json:"ip"`
	SubType      string             `bson:"subType" json:"subType"`
	WireGuard    WireGuard          `bson:"wireguard" json:"wireguard"`
	Subscription *Subscription      `bson:"subscription,omitempty" json:"subscription,omitempty"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
}
