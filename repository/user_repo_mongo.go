package repository

import (
	"context"
	"time"

	"hariomtransport/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoUserRepo struct {
	DB *mongo.Client
}

func NewMongoUserRepo(db *mongo.Client) *MongoUserRepo {
	return &MongoUserRepo{DB: db}
}

func (r *MongoUserRepo) CreateUser(user *models.AppUser) error {
	ctx := context.Background()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}

	_, err := r.DB.Database("hariomtransport").Collection("app_user").InsertOne(ctx, user)
	return err
}

func (r *MongoUserRepo) GetUserByEmail(email string) (*models.AppUser, error) {
	ctx := context.Background()
	user := &models.AppUser{}

	err := r.DB.Database("hariomtransport").Collection("app_user").
		FindOne(ctx, bson.M{"email": email}).Decode(user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}
