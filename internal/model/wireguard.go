package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type WireGuard struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PrivateKey string             `bson:"privateKey" json:"privateKey"`
	PublicKey  string             `bson:"publicKey" json:"publicKey"`
	IP         string             `bson:"ip" json:"ip"`
}
