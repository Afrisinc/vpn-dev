package repository

import (
	"context"

	"github.com/ElissaDesign/vpn-dev/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

func (r *UserRepository) GetAll(ctx context.Context) ([]model.User, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []model.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) (*mongo.InsertOneResult, error) {
	return r.collection.InsertOne(ctx, user)
}
