package db

import (
	"context"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var clientInstance *mongo.Client
var clientOnce sync.Once

func GetMongoClient() *mongo.Client {
	clientOnce.Do(func() {
		uri := ""

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		clientInstance, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
		if err != nil {
			log.Fatalf("MongoDB connection error: %v", err)
		}

		if err := clientInstance.Ping(ctx, nil); err != nil {
			log.Fatalf("MongoDB ping error: %v", err)
		}

		log.Println("Connected to MongoDB")
	})

	return clientInstance
}
