package repository

import (
	"context"
	"time"

	"hariomtransport/models"

	"go.mongodb.org/mongo-driver/mongo"
)

type MongoInitialRepo struct {
	DB *mongo.Client
}

func NewMongoInitialRepo(db *mongo.Client) *MongoInitialRepo {
	return &MongoInitialRepo{DB: db}
}

func (r *MongoInitialRepo) SaveInitial(initial *models.InitialSetup) error {
	ctx := context.Background()
	db := r.DB.Database("hariomtransport")

	if initial.CreatedAt.IsZero() {
		initial.CreatedAt = time.Now().UTC()
	}

	_, err := db.Collection("initial_setup").InsertOne(ctx, initial)
	return err
}

func (r *MongoInitialRepo) GetInitial() (*models.InitialSetup, error) {
	ctx := context.Background()
	db := r.DB.Database("hariomtransport")

	var initial models.InitialSetup
	err := db.Collection("initial_setup").FindOne(ctx, struct{}{}).Decode(&initial)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &initial, nil
}
