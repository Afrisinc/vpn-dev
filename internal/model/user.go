package model

// import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
